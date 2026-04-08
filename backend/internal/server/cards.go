package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const defaultPageLimit = 20
const maxPageLimit = 100

type cardResponse struct {
	ID        string `json:"id"`
	Issuer    string `json:"issuer"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	AnnualFee int    `json:"annual_fee"`
	ImageURL  string `json:"image_url"`
	Active    bool   `json:"active"`
}

type creditDefResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AmountCents int    `json:"amount_cents"`
	Period      string `json:"period"`
	Category    string `json:"category"`
}

type cardDetailResponse struct {
	cardResponse
	Credits []creditDefResponse `json:"credits"`
}

func (s *Server) handleListCards(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := parseLimit(q.Get("limit"))
	cursor := q.Get("cursor")
	issuer := q.Get("issuer")
	search := q.Get("q")

	// Build query dynamically based on filters.
	query := `SELECT id, issuer, name, network, annual_fee, image_url, active
	          FROM card_catalog WHERE active = true`
	args := []any{}
	argN := 1

	if issuer != "" {
		query += ` AND issuer = $` + strconv.Itoa(argN)
		args = append(args, issuer)
		argN++
	}
	if search != "" {
		query += ` AND (name ILIKE $` + strconv.Itoa(argN) + ` OR issuer ILIKE $` + strconv.Itoa(argN) + `)`
		args = append(args, "%"+search+"%")
		argN++
	}
	if cursor != "" {
		query += ` AND id > $` + strconv.Itoa(argN)
		args = append(args, cursor)
		argN++
	}

	query += ` ORDER BY id LIMIT $` + strconv.Itoa(argN)
	args = append(args, limit+1) // fetch one extra to determine next_cursor

	rows, err := s.db.Query(r.Context(), query, args...)
	if err != nil {
		s.logger.Error("list cards: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	cards := []cardResponse{}
	for rows.Next() {
		var c cardResponse
		if err := rows.Scan(&c.ID, &c.Issuer, &c.Name, &c.Network, &c.AnnualFee, &c.ImageURL, &c.Active); err != nil {
			s.logger.Error("list cards: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		cards = append(cards, c)
	}

	var nextCursor *string
	if len(cards) > limit {
		nc := cards[limit].ID
		nextCursor = &nc
		cards = cards[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":        cards,
		"next_cursor": nextCursor,
	})
}

func (s *Server) handleGetCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "card_id")

	var c cardResponse
	err := s.db.QueryRow(r.Context(),
		`SELECT id, issuer, name, network, annual_fee, image_url, active
		 FROM card_catalog WHERE id = $1`, cardID,
	).Scan(&c.ID, &c.Issuer, &c.Name, &c.Network, &c.AnnualFee, &c.ImageURL, &c.Active)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "card not found"})
		return
	}

	rows, err := s.db.Query(r.Context(),
		`SELECT id, name, description, amount_cents, period, category
		 FROM credit_definitions WHERE card_catalog_id = $1
		 ORDER BY period, name`, cardID,
	)
	if err != nil {
		s.logger.Error("get card: query credits", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	credits := []creditDefResponse{}
	for rows.Next() {
		var cd creditDefResponse
		if err := rows.Scan(&cd.ID, &cd.Name, &cd.Description, &cd.AmountCents, &cd.Period, &cd.Category); err != nil {
			s.logger.Error("get card: scan credit", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		credits = append(credits, cd)
	}

	writeJSON(w, http.StatusOK, cardDetailResponse{
		cardResponse: c,
		Credits:      credits,
	})
}

func parseLimit(s string) int {
	if s == "" {
		return defaultPageLimit
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultPageLimit
	}
	if n > maxPageLimit {
		return maxPageLimit
	}
	return n
}
