package server

import (
	"net/http"
	"strconv"
	"time"
)

type cardSummary struct {
	UserCardID    string `json:"user_card_id"`
	CardName      string `json:"card_name"`
	Issuer        string `json:"issuer"`
	AvailableCents int   `json:"available_cents"`
	UsedCents     int    `json:"used_cents"`
	CreditsUsed   int    `json:"credits_used"`
	CreditsTotal  int    `json:"credits_total"`
}

type dashboardSummary struct {
	TotalAvailableCents int           `json:"total_available_cents"`
	TotalUsedCents      int           `json:"total_used_cents"`
	TotalUnusedCents    int           `json:"total_unused_cents"`
	CreditsUsed         int           `json:"credits_used"`
	CreditsTotal        int           `json:"credits_total"`
	Cards               []cardSummary `json:"cards"`
}

type monthEntry struct {
	Month          int `json:"month"`
	AvailableCents int `json:"available_cents"`
	UsedCents      int `json:"used_cents"`
}

type annualSummary struct {
	Year                int          `json:"year"`
	TotalAvailableCents int          `json:"total_available_cents"`
	TotalUsedCents      int          `json:"total_used_cents"`
	Months              []monthEntry `json:"months"`
}

type monthlySummary struct {
	Year                int           `json:"year"`
	Month               int           `json:"month"`
	TotalAvailableCents int           `json:"total_available_cents"`
	TotalUsedCents      int           `json:"total_used_cents"`
	Cards               []cardSummary `json:"cards"`
}

// handleDashboardSummary returns aggregated credit usage for all current periods.
func (s *Server) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	today := time.Now().Format("2006-01-02")

	rows, err := s.db.Query(r.Context(),
		`SELECT uc.id, cc.name, cc.issuer,
		        cd.amount_cents, cp.used, cp.amount_used_cents
		 FROM credit_periods cp
		 JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
		 JOIN user_cards uc ON uc.id = cp.user_card_id
		 JOIN card_catalog cc ON cc.id = uc.card_catalog_id
		 WHERE uc.user_id = $1
		   AND cp.period_start <= $2 AND cp.period_end >= $2
		 ORDER BY uc.id`, userID, today,
	)
	if err != nil {
		s.logger.Error("dashboard summary: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	cardMap := map[string]*cardSummary{}
	cardOrder := []string{}
	summary := dashboardSummary{Cards: []cardSummary{}}

	for rows.Next() {
		var ucID, cardName, issuer string
		var amountCents, amountUsedCents int
		var used bool

		if err := rows.Scan(&ucID, &cardName, &issuer, &amountCents, &used, &amountUsedCents); err != nil {
			s.logger.Error("dashboard summary: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		cs, exists := cardMap[ucID]
		if !exists {
			cs = &cardSummary{UserCardID: ucID, CardName: cardName, Issuer: issuer}
			cardMap[ucID] = cs
			cardOrder = append(cardOrder, ucID)
		}

		cs.AvailableCents += amountCents
		cs.CreditsTotal++
		summary.TotalAvailableCents += amountCents
		summary.CreditsTotal++

		if used {
			cs.UsedCents += amountUsedCents
			cs.CreditsUsed++
			summary.TotalUsedCents += amountUsedCents
			summary.CreditsUsed++
		}
	}

	summary.TotalUnusedCents = summary.TotalAvailableCents - summary.TotalUsedCents

	for _, id := range cardOrder {
		summary.Cards = append(summary.Cards, *cardMap[id])
	}

	writeJSON(w, http.StatusOK, summary)
}

// handleDashboardAnnual returns a monthly breakdown for a given year.
func (s *Server) handleDashboardAnnual(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	year := time.Now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}

	// Get all credit periods that overlap with any month in the year.
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	yearEnd := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	rows, err := s.db.Query(r.Context(),
		`SELECT cd.amount_cents, cp.used, cp.amount_used_cents,
		        cp.period_start, cp.period_end
		 FROM credit_periods cp
		 JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
		 JOIN user_cards uc ON uc.id = cp.user_card_id
		 WHERE uc.user_id = $1
		   AND cp.period_end >= $2 AND cp.period_start <= $3
		 ORDER BY cp.period_start`, userID, yearStart, yearEnd,
	)
	if err != nil {
		s.logger.Error("dashboard annual: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	// Initialize all 12 months.
	months := make([]monthEntry, 12)
	for i := range months {
		months[i].Month = i + 1
	}

	var totalAvail, totalUsed int

	for rows.Next() {
		var amountCents, amountUsedCents int
		var used bool
		var periodStart, periodEnd time.Time

		if err := rows.Scan(&amountCents, &used, &amountUsedCents, &periodStart, &periodEnd); err != nil {
			s.logger.Error("dashboard annual: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// Attribute to the month containing the period midpoint.
		mid := periodStart.Add(periodEnd.Sub(periodStart) / 2)
		if mid.Year() == year {
			m := int(mid.Month()) - 1
			months[m].AvailableCents += amountCents
			totalAvail += amountCents
			if used {
				months[m].UsedCents += amountUsedCents
				totalUsed += amountUsedCents
			}
		}
	}

	writeJSON(w, http.StatusOK, annualSummary{
		Year:                year,
		TotalAvailableCents: totalAvail,
		TotalUsedCents:      totalUsed,
		Months:              months,
	})
}

// handleDashboardMonthly returns per-card breakdown for a specific month.
func (s *Server) handleDashboardMonthly(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	now := time.Now()

	year := now.Year()
	month := int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	// Find credit periods whose midpoint falls in this month.
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	monthEnd := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	rows, err := s.db.Query(r.Context(),
		`SELECT uc.id, cc.name, cc.issuer,
		        cd.amount_cents, cp.used, cp.amount_used_cents,
		        cp.period_start, cp.period_end
		 FROM credit_periods cp
		 JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
		 JOIN user_cards uc ON uc.id = cp.user_card_id
		 JOIN card_catalog cc ON cc.id = uc.card_catalog_id
		 WHERE uc.user_id = $1
		   AND cp.period_end >= $2 AND cp.period_start <= $3
		 ORDER BY uc.id`, userID, monthStart, monthEnd,
	)
	if err != nil {
		s.logger.Error("dashboard monthly: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	cardMap := map[string]*cardSummary{}
	cardOrder := []string{}
	var totalAvail, totalUsed int

	for rows.Next() {
		var ucID, cardName, issuer string
		var amountCents, amountUsedCents int
		var used bool
		var periodStart, periodEnd time.Time

		if err := rows.Scan(&ucID, &cardName, &issuer, &amountCents, &used, &amountUsedCents, &periodStart, &periodEnd); err != nil {
			s.logger.Error("dashboard monthly: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// Only count if period midpoint is in this month.
		mid := periodStart.Add(periodEnd.Sub(periodStart) / 2)
		if mid.Year() != year || int(mid.Month()) != month {
			continue
		}

		cs, exists := cardMap[ucID]
		if !exists {
			cs = &cardSummary{UserCardID: ucID, CardName: cardName, Issuer: issuer}
			cardMap[ucID] = cs
			cardOrder = append(cardOrder, ucID)
		}

		cs.AvailableCents += amountCents
		cs.CreditsTotal++
		totalAvail += amountCents

		if used {
			cs.UsedCents += amountUsedCents
			cs.CreditsUsed++
			totalUsed += amountUsedCents
		}
	}

	cards := make([]cardSummary, 0, len(cardOrder))
	for _, id := range cardOrder {
		cards = append(cards, *cardMap[id])
	}

	writeJSON(w, http.StatusOK, monthlySummary{
		Year:                year,
		Month:               month,
		TotalAvailableCents: totalAvail,
		TotalUsedCents:      totalUsed,
		Cards:               cards,
	})
}
