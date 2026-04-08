package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"creditcatch/backend/internal/auth"
)

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         authUser `json:"user"`
}

type authUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("signup: hash password", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	var userID, email, name string
	err = s.db.QueryRow(r.Context(),
		`INSERT INTO users (email, password, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name`,
		req.Email, hash, req.Name,
	).Scan(&userID, &email, &name)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		s.logger.Error("signup: insert user", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	s.issueTokenPair(w, r, userID, email, name, http.StatusCreated)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var userID, hash, email, name string
	err := s.db.QueryRow(r.Context(),
		`SELECT id, password, email, name FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &hash, &email, &name)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if hash == "" || !auth.CheckPassword(hash, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	s.issueTokenPair(w, r, userID, email, name, http.StatusOK)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	newRefresh, userID, err := auth.RotateRefreshToken(r.Context(), s.db, req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenRevoked) {
			s.logger.Warn("refresh: revoked token reuse detected", "error", err)
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		return
	}

	var email, name string
	err = s.db.QueryRow(r.Context(),
		`SELECT email, name FROM users WHERE id = $1`, userID,
	).Scan(&email, &name)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	accessToken, err := s.jwt.Issue(userID)
	if err != nil {
		s.logger.Error("refresh: issue access token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		User:         authUser{ID: userID, Email: email, Name: name},
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
		return
	}

	if req.RefreshToken != "" {
		if err := auth.RevokeFamily(r.Context(), s.db, req.RefreshToken); err != nil {
			s.logger.Error("logout: revoke refresh token family", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// issueTokenPair creates a new access token + refresh token and writes the response.
func (s *Server) issueTokenPair(w http.ResponseWriter, r *http.Request, userID, email, name string, status int) {
	accessToken, err := s.jwt.Issue(userID)
	if err != nil {
		s.logger.Error("issue token pair: access token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	family := uuid.New().String()
	refreshToken, err := auth.GenerateRefreshToken(r.Context(), s.db, userID, family)
	if err != nil {
		s.logger.Error("issue token pair: refresh token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, status, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         authUser{ID: userID, Email: email, Name: name},
	})
}
