package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"creditcatch/backend/internal/credits"

	"github.com/go-chi/chi/v5"
)

type addUserCardRequest struct {
	CardCatalogID     string  `json:"card_catalog_id"`
	Nickname          string  `json:"nickname"`
	OpenedDate        *string `json:"opened_date"`
	AnnualFeeDate     *string `json:"annual_fee_date"`
	StatementCloseDay *int    `json:"statement_close_day"`
}

type updateUserCardRequest struct {
	Nickname          *string `json:"nickname"`
	OpenedDate        *string `json:"opened_date"`
	AnnualFeeDate     *string `json:"annual_fee_date"`
	StatementCloseDay *int    `json:"statement_close_day"`
	Active            *bool   `json:"active"`
}

type userCardResponse struct {
	ID                string     `json:"id"`
	Card              cardResponse `json:"card"`
	Nickname          string     `json:"nickname"`
	OpenedDate        *string    `json:"opened_date"`
	AnnualFeeDate     *string    `json:"annual_fee_date"`
	StatementCloseDay *int       `json:"statement_close_day"`
	Active            bool       `json:"active"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (s *Server) handleListUserCards(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	rows, err := s.db.Query(r.Context(),
		`SELECT uc.id, uc.nickname, uc.opened_date, uc.annual_fee_date,
		        uc.statement_close_day, uc.active, uc.created_at,
		        cc.id, cc.issuer, cc.name, cc.network, cc.annual_fee, cc.image_url, cc.active
		 FROM user_cards uc
		 JOIN card_catalog cc ON cc.id = uc.card_catalog_id
		 WHERE uc.user_id = $1
		 ORDER BY uc.created_at DESC`, userID,
	)
	if err != nil {
		s.logger.Error("list user cards: query", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	cards := []userCardResponse{}
	for rows.Next() {
		var uc userCardResponse
		if err := rows.Scan(
			&uc.ID, &uc.Nickname, &uc.OpenedDate, &uc.AnnualFeeDate,
			&uc.StatementCloseDay, &uc.Active, &uc.CreatedAt,
			&uc.Card.ID, &uc.Card.Issuer, &uc.Card.Name, &uc.Card.Network,
			&uc.Card.AnnualFee, &uc.Card.ImageURL, &uc.Card.Active,
		); err != nil {
			s.logger.Error("list user cards: scan", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		cards = append(cards, uc)
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": cards})
}

func (s *Server) handleAddUserCard(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	var req addUserCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.CardCatalogID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "card_catalog_id is required"})
		return
	}
	if req.StatementCloseDay == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "statement_close_day is required"})
		return
	}
	if *req.StatementCloseDay < 1 || *req.StatementCloseDay > 31 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "statement_close_day must be between 1 and 31"})
		return
	}

	var ucID string
	err := s.db.QueryRow(r.Context(),
		`INSERT INTO user_cards (user_id, card_catalog_id, nickname, opened_date, annual_fee_date, statement_close_day)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		userID, req.CardCatalogID, req.Nickname, req.OpenedDate, req.AnnualFeeDate, req.StatementCloseDay,
	).Scan(&ucID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "you already have this card"})
			return
		}
		if strings.Contains(err.Error(), "foreign key") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "card not found in catalog"})
			return
		}
		s.logger.Error("add user card: insert", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Generate credit periods for this card.
	var openedDate *time.Time
	if req.OpenedDate != nil {
		if t, err := time.Parse("2006-01-02", *req.OpenedDate); err == nil {
			openedDate = &t
		}
	}
	if err := credits.GeneratePeriodsForCard(r.Context(), s.db, ucID, *req.StatementCloseDay, openedDate); err != nil {
		s.logger.Error("add user card: generate periods", "error", err)
		// Non-fatal: card was created, periods can be regenerated.
	}

	// Fetch the full response with joined card data.
	uc, err := s.fetchUserCard(r, ucID, userID)
	if err != nil {
		s.logger.Error("add user card: fetch", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, uc)
}

func (s *Server) handleGetUserCard(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ucID := chi.URLParam(r, "user_card_id")

	uc, err := s.fetchUserCard(r, ucID, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "card not found"})
		return
	}

	writeJSON(w, http.StatusOK, uc)
}

func (s *Server) handleUpdateUserCard(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ucID := chi.URLParam(r, "user_card_id")

	var req updateUserCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.StatementCloseDay != nil && (*req.StatementCloseDay < 1 || *req.StatementCloseDay > 31) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "statement_close_day must be between 1 and 31"})
		return
	}

	// Build dynamic UPDATE.
	sets := []string{}
	args := []any{}
	argN := 1

	if req.Nickname != nil {
		sets = append(sets, "nickname = $"+strconv.Itoa(argN))
		args = append(args, *req.Nickname)
		argN++
	}
	if req.OpenedDate != nil {
		sets = append(sets, "opened_date = $"+strconv.Itoa(argN))
		args = append(args, *req.OpenedDate)
		argN++
	}
	if req.AnnualFeeDate != nil {
		sets = append(sets, "annual_fee_date = $"+strconv.Itoa(argN))
		args = append(args, *req.AnnualFeeDate)
		argN++
	}
	if req.StatementCloseDay != nil {
		sets = append(sets, "statement_close_day = $"+strconv.Itoa(argN))
		args = append(args, *req.StatementCloseDay)
		argN++
	}
	if req.Active != nil {
		sets = append(sets, "active = $"+strconv.Itoa(argN))
		args = append(args, *req.Active)
		argN++
	}

	if len(sets) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no fields to update"})
		return
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE user_cards SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(argN) + " AND user_id = $" + strconv.Itoa(argN+1)
	args = append(args, ucID, userID)

	tag, err := s.db.Exec(r.Context(), query, args...)
	if err != nil {
		s.logger.Error("update user card: exec", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "card not found"})
		return
	}

	uc, err := s.fetchUserCard(r, ucID, userID)
	if err != nil {
		s.logger.Error("update user card: fetch", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, uc)
}

func (s *Server) handleDeleteUserCard(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ucID := chi.URLParam(r, "user_card_id")

	tag, err := s.db.Exec(r.Context(),
		`DELETE FROM user_cards WHERE id = $1 AND user_id = $2`, ucID, userID,
	)
	if err != nil {
		s.logger.Error("delete user card: exec", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "card not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// fetchUserCard loads a single user card with joined catalog data.
func (s *Server) fetchUserCard(r *http.Request, ucID, userID string) (*userCardResponse, error) {
	var uc userCardResponse
	err := s.db.QueryRow(r.Context(),
		`SELECT uc.id, uc.nickname, uc.opened_date, uc.annual_fee_date,
		        uc.statement_close_day, uc.active, uc.created_at,
		        cc.id, cc.issuer, cc.name, cc.network, cc.annual_fee, cc.image_url, cc.active
		 FROM user_cards uc
		 JOIN card_catalog cc ON cc.id = uc.card_catalog_id
		 WHERE uc.id = $1 AND uc.user_id = $2`, ucID, userID,
	).Scan(
		&uc.ID, &uc.Nickname, &uc.OpenedDate, &uc.AnnualFeeDate,
		&uc.StatementCloseDay, &uc.Active, &uc.CreatedAt,
		&uc.Card.ID, &uc.Card.Issuer, &uc.Card.Name, &uc.Card.Network,
		&uc.Card.AnnualFee, &uc.Card.ImageURL, &uc.Card.Active,
	)
	if err != nil {
		return nil, err
	}
	return &uc, nil
}

