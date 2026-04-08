package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type creditPeriodResponse struct {
	ID               string             `json:"id"`
	UserCardID       string             `json:"user_card_id"`
	CreditDefinition creditDefResponse  `json:"credit_definition"`
	PeriodStart      string             `json:"period_start"`
	PeriodEnd        string             `json:"period_end"`
	Used             bool               `json:"used"`
	UsedAt           *time.Time         `json:"used_at"`
	AmountUsedCents  int                `json:"amount_used_cents"`
	TransactionID    *string            `json:"transaction_id"`
}

type currentCreditResponse struct {
	UserCard userCardResponse       `json:"user_card"`
	Credits  []creditPeriodResponse `json:"credits"`
}

type markUsedRequest struct {
	AmountUsedCents *int    `json:"amount_used_cents"`
	TransactionID   *string `json:"transaction_id"`
}

func (s *Server) handleListCredits(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	q := r.URL.Query()
	limit := parseLimit(q.Get("limit"))
	cursor := q.Get("cursor")
	status := q.Get("status")
	userCardID := q.Get("user_card_id")

	query := `SELECT cp.id, cp.user_card_id, cp.period_start, cp.period_end,
	                 cp.used, cp.used_at, cp.amount_used_cents, cp.transaction_id,
	                 cd.id, cd.name, cd.description, cd.amount_cents, cd.period, cd.category
	          FROM credit_periods cp
	          JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
	          JOIN user_cards uc ON uc.id = cp.user_card_id
	          WHERE uc.user_id = $1`
	args := []any{userID}
	argN := 2

	if status == "used" {
		query += ` AND cp.used = true`
	} else if status == "unused" {
		query += ` AND cp.used = false`
	}

	if userCardID != "" {
		query += ` AND cp.user_card_id = $` + strconv.Itoa(argN)
		args = append(args, userCardID)
		argN++
	}

	if cursor != "" {
		query += ` AND cp.id > $` + strconv.Itoa(argN)
		args = append(args, cursor)
		argN++
	}

	query += ` ORDER BY cp.period_start DESC, cp.id LIMIT $` + strconv.Itoa(argN)
	args = append(args, limit+1)

	rows, err := s.db.Query(r.Context(), query, args...)
	if err != nil {
		s.logger.Error("list credits: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	periods := []creditPeriodResponse{}
	for rows.Next() {
		var p creditPeriodResponse
		if err := rows.Scan(
			&p.ID, &p.UserCardID, &p.PeriodStart, &p.PeriodEnd,
			&p.Used, &p.UsedAt, &p.AmountUsedCents, &p.TransactionID,
			&p.CreditDefinition.ID, &p.CreditDefinition.Name, &p.CreditDefinition.Description,
			&p.CreditDefinition.AmountCents, &p.CreditDefinition.Period, &p.CreditDefinition.Category,
		); err != nil {
			s.logger.Error("list credits: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		periods = append(periods, p)
	}

	var nextCursor *string
	if len(periods) > limit {
		nc := periods[limit].ID
		nextCursor = &nc
		periods = periods[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":        periods,
		"next_cursor": nextCursor,
	})
}

func (s *Server) handleCurrentCredits(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	today := time.Now().Format("2006-01-02")

	// Get all current credit periods grouped by user card.
	rows, err := s.db.Query(r.Context(),
		`SELECT cp.id, cp.user_card_id, cp.period_start, cp.period_end,
		        cp.used, cp.used_at, cp.amount_used_cents, cp.transaction_id,
		        cd.id, cd.name, cd.description, cd.amount_cents, cd.period, cd.category,
		        uc.id, uc.nickname, uc.opened_date, uc.annual_fee_date,
		        uc.statement_close_day, uc.active, uc.created_at,
		        cc.id, cc.issuer, cc.name, cc.network, cc.annual_fee, cc.image_url, cc.active
		 FROM credit_periods cp
		 JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
		 JOIN user_cards uc ON uc.id = cp.user_card_id
		 JOIN card_catalog cc ON cc.id = uc.card_catalog_id
		 WHERE uc.user_id = $1
		   AND cp.period_start <= $2 AND cp.period_end >= $2
		 ORDER BY uc.id, cd.name`, userID, today,
	)
	if err != nil {
		s.logger.Error("current credits: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	// Group by user card.
	grouped := map[string]*currentCreditResponse{}
	order := []string{}

	for rows.Next() {
		var p creditPeriodResponse
		var uc userCardResponse

		if err := rows.Scan(
			&p.ID, &p.UserCardID, &p.PeriodStart, &p.PeriodEnd,
			&p.Used, &p.UsedAt, &p.AmountUsedCents, &p.TransactionID,
			&p.CreditDefinition.ID, &p.CreditDefinition.Name, &p.CreditDefinition.Description,
			&p.CreditDefinition.AmountCents, &p.CreditDefinition.Period, &p.CreditDefinition.Category,
			&uc.ID, &uc.Nickname, &uc.OpenedDate, &uc.AnnualFeeDate,
			&uc.StatementCloseDay, &uc.Active, &uc.CreatedAt,
			&uc.Card.ID, &uc.Card.Issuer, &uc.Card.Name, &uc.Card.Network,
			&uc.Card.AnnualFee, &uc.Card.ImageURL, &uc.Card.Active,
		); err != nil {
			s.logger.Error("current credits: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if _, exists := grouped[uc.ID]; !exists {
			grouped[uc.ID] = &currentCreditResponse{
				UserCard: uc,
				Credits:  []creditPeriodResponse{},
			}
			order = append(order, uc.ID)
		}
		grouped[uc.ID].Credits = append(grouped[uc.ID].Credits, p)
	}

	result := make([]currentCreditResponse, 0, len(order))
	for _, id := range order {
		result = append(result, *grouped[id])
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (s *Server) handleMarkUsed(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	cpID := chi.URLParam(r, "credit_period_id")

	var req markUsedRequest
	// Body is optional per spec.
	json.NewDecoder(r.Body).Decode(&req)

	// Determine amount: use provided amount or default to full credit amount.
	amountQuery := `COALESCE($3, cd.amount_cents)`
	now := time.Now()

	tag, err := s.db.Exec(r.Context(),
		`UPDATE credit_periods cp
		 SET used = true, used_at = $4, amount_used_cents = `+amountQuery+`, transaction_id = $5, updated_at = now()
		 FROM credit_definitions cd, user_cards uc
		 WHERE cp.id = $1
		   AND cp.credit_definition_id = cd.id
		   AND cp.user_card_id = uc.id
		   AND uc.user_id = $2`,
		cpID, userID, req.AmountUsedCents, now, req.TransactionID,
	)
	if err != nil {
		s.logger.Error("mark used: exec", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "credit period not found"})
		return
	}

	s.respondWithCreditPeriod(w, r, cpID, userID)
}

func (s *Server) handleMarkUnused(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	cpID := chi.URLParam(r, "credit_period_id")

	tag, err := s.db.Exec(r.Context(),
		`UPDATE credit_periods cp
		 SET used = false, used_at = NULL, amount_used_cents = 0, transaction_id = NULL, updated_at = now()
		 FROM user_cards uc
		 WHERE cp.id = $1
		   AND cp.user_card_id = uc.id
		   AND uc.user_id = $2`,
		cpID, userID,
	)
	if err != nil {
		s.logger.Error("mark unused: exec", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "credit period not found"})
		return
	}

	s.respondWithCreditPeriod(w, r, cpID, userID)
}

func (s *Server) respondWithCreditPeriod(w http.ResponseWriter, r *http.Request, cpID, userID string) {
	var p creditPeriodResponse
	err := s.db.QueryRow(r.Context(),
		`SELECT cp.id, cp.user_card_id, cp.period_start, cp.period_end,
		        cp.used, cp.used_at, cp.amount_used_cents, cp.transaction_id,
		        cd.id, cd.name, cd.description, cd.amount_cents, cd.period, cd.category
		 FROM credit_periods cp
		 JOIN credit_definitions cd ON cd.id = cp.credit_definition_id
		 JOIN user_cards uc ON uc.id = cp.user_card_id
		 WHERE cp.id = $1 AND uc.user_id = $2`, cpID, userID,
	).Scan(
		&p.ID, &p.UserCardID, &p.PeriodStart, &p.PeriodEnd,
		&p.Used, &p.UsedAt, &p.AmountUsedCents, &p.TransactionID,
		&p.CreditDefinition.ID, &p.CreditDefinition.Name, &p.CreditDefinition.Description,
		&p.CreditDefinition.AmountCents, &p.CreditDefinition.Period, &p.CreditDefinition.Category,
	)
	if err != nil {
		s.logger.Error("fetch credit period", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, p)
}
