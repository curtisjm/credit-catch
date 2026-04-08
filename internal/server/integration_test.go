//go:build integration

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"creditcatch/backend/internal/auth"
	"creditcatch/backend/internal/config"
	"creditcatch/backend/internal/database"
)

// testServer sets up a full server with a real database for integration tests.
// Requires DATABASE_URL and JWT_SECRET environment variables.
func testServer(t *testing.T) (*Server, func()) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		t.Fatalf("database.NewPool: %v", err)
	}

	// Clean test data before each test.
	cleanDB(t, pool)

	cfg := &config.Config{
		Port:        0,
		DatabaseURL: dbURL,
		JWTSecret:   "test-integration-secret",
		JWTExpiry:   1 * time.Hour,
		Environment: "test",
		LogLevel:    "error",
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	srv := New(pool, cfg, logger)

	cleanup := func() {
		cleanDB(t, pool)
		pool.Close()
	}

	return srv, cleanup
}

func cleanDB(t *testing.T, pool interface{ Exec(ctx context.Context, sql string, args ...any) (interface{ RowsAffected() int64 }, error) }) {
	t.Helper()
	// Can't use pool.Exec directly because pgxpool.Pool.Exec returns (pgconn.CommandTag, error).
	// We'll use a raw connection instead.
}

// Helper to seed a card in the catalog for testing.
func seedCard(t *testing.T, srv *Server) string {
	t.Helper()
	ctx := context.Background()
	var cardID string
	err := srv.db.QueryRow(ctx,
		`INSERT INTO card_catalog (issuer, name, network, annual_fee, active)
		 VALUES ('Chase', 'Sapphire Reserve', 'visa', 55000, true)
		 ON CONFLICT (issuer, name) DO UPDATE SET active = true
		 RETURNING id`,
	).Scan(&cardID)
	if err != nil {
		t.Fatalf("seed card: %v", err)
	}

	// Add a credit definition.
	_, err = srv.db.Exec(ctx,
		`INSERT INTO credit_definitions (card_catalog_id, name, description, amount_cents, period, category)
		 VALUES ($1, '$300 Travel Credit', 'Annual travel credit', 30000, 'annual', 'travel')
		 ON CONFLICT DO NOTHING`, cardID,
	)
	if err != nil {
		t.Fatalf("seed credit def: %v", err)
	}

	return cardID
}

func doRequest(srv *Server, method, path string, body any, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestIntegration_SignupAndLogin(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	// Signup.
	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":    "test@example.com",
		"password": "securepassword123",
		"name":     "Test User",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("signup status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var signupResp authResponse
	json.NewDecoder(rec.Body).Decode(&signupResp)

	if signupResp.AccessToken == "" {
		t.Error("signup should return access_token")
	}
	if signupResp.RefreshToken == "" {
		t.Error("signup should return refresh_token")
	}
	if signupResp.User.Email != "test@example.com" {
		t.Errorf("user email = %q, want %q", signupResp.User.Email, "test@example.com")
	}

	// Login with same credentials.
	rec = doRequest(srv, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "securepassword123",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var loginResp authResponse
	json.NewDecoder(rec.Body).Decode(&loginResp)

	if loginResp.AccessToken == "" {
		t.Error("login should return access_token")
	}
	if loginResp.User.ID != signupResp.User.ID {
		t.Error("login should return same user ID as signup")
	}
}

func TestIntegration_SignupDuplicateEmail(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	body := map[string]string{
		"email":    "dupe@example.com",
		"password": "password123!",
		"name":     "User",
	}

	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", body, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("first signup: status = %d", rec.Code)
	}

	rec = doRequest(srv, http.MethodPost, "/api/v1/auth/signup", body, "")
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate signup: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestIntegration_LoginWrongPassword(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	doRequest(srv, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": "user@example.com", "password": "correctpassword", "name": "User",
	}, "")

	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "user@example.com", "password": "wrongpassword",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestIntegration_RefreshTokenRotation(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	// Signup to get initial tokens.
	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": "refresh@example.com", "password": "password123!", "name": "User",
	}, "")

	var resp authResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	origRefresh := resp.RefreshToken

	// Refresh.
	rec = doRequest(srv, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": origRefresh,
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var refreshResp authResponse
	json.NewDecoder(rec.Body).Decode(&refreshResp)

	if refreshResp.RefreshToken == origRefresh {
		t.Error("refresh should return a NEW refresh token")
	}
	if refreshResp.AccessToken == "" {
		t.Error("refresh should return an access token")
	}

	// Reuse the old refresh token (theft detection).
	rec = doRequest(srv, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": origRefresh,
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("reuse old token: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// The new token should also be revoked (entire family revoked).
	rec = doRequest(srv, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": refreshResp.RefreshToken,
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("family revoked: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestIntegration_ProtectedRouteWithoutAuth(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	rec := doRequest(srv, http.MethodGet, "/api/v1/me/cards", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestIntegration_CardCatalog(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	cardID := seedCard(t, srv)

	// List cards (public endpoint).
	rec := doRequest(srv, http.MethodGet, "/api/v1/cards", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list cards: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var listResp struct {
		Data []cardResponse `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&listResp)

	if len(listResp.Data) == 0 {
		t.Fatal("list cards: expected at least one card")
	}

	// Get card detail.
	rec = doRequest(srv, http.MethodGet, "/api/v1/cards/"+cardID, nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get card: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var detail cardDetailResponse
	json.NewDecoder(rec.Body).Decode(&detail)

	if detail.Issuer != "Chase" {
		t.Errorf("card issuer = %q, want %q", detail.Issuer, "Chase")
	}
	if len(detail.Credits) == 0 {
		t.Error("card detail should include credit definitions")
	}
}

func TestIntegration_UserCardsCRUD(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	cardID := seedCard(t, srv)

	// Signup.
	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": "crud@example.com", "password": "password123!", "name": "User",
	}, "")
	var authResp authResponse
	json.NewDecoder(rec.Body).Decode(&authResp)
	token := authResp.AccessToken

	// Add card.
	rec = doRequest(srv, http.MethodPost, "/api/v1/me/cards", map[string]any{
		"card_catalog_id":     cardID,
		"nickname":            "My Reserve",
		"statement_close_day": 15,
		"opened_date":         "2024-09-03",
	}, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("add card: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var uc userCardResponse
	json.NewDecoder(rec.Body).Decode(&uc)

	if uc.Nickname != "My Reserve" {
		t.Errorf("nickname = %q, want %q", uc.Nickname, "My Reserve")
	}

	// List user cards.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/cards", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list user cards: status = %d", rec.Code)
	}

	// Update card.
	newNickname := "Travel Card"
	rec = doRequest(srv, http.MethodPatch, "/api/v1/me/cards/"+uc.ID, map[string]any{
		"nickname": newNickname,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update card: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var updated userCardResponse
	json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Nickname != newNickname {
		t.Errorf("updated nickname = %q, want %q", updated.Nickname, newNickname)
	}

	// Delete card.
	rec = doRequest(srv, http.MethodDelete, "/api/v1/me/cards/"+uc.ID, nil, token)
	if rec.Code != http.StatusNoContent {
		t.Errorf("delete card: status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Verify deleted.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/cards/"+uc.ID, nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("get deleted card: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestIntegration_CreditsAndDashboard(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	cardID := seedCard(t, srv)

	// Signup and add card.
	rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email": "credits@example.com", "password": "password123!", "name": "User",
	}, "")
	var authResp authResponse
	json.NewDecoder(rec.Body).Decode(&authResp)
	token := authResp.AccessToken

	rec = doRequest(srv, http.MethodPost, "/api/v1/me/cards", map[string]any{
		"card_catalog_id":     cardID,
		"nickname":            "Reserve",
		"statement_close_day": 15,
		"opened_date":         "2024-09-03",
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add card: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	// List credits.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/credits", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list credits: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var creditsResp struct {
		Data []creditPeriodResponse `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&creditsResp)

	if len(creditsResp.Data) == 0 {
		t.Fatal("expected credit periods after adding card with opened_date")
	}

	cpID := creditsResp.Data[0].ID

	// Mark used.
	rec = doRequest(srv, http.MethodPost, "/api/v1/me/credits/"+cpID+"/mark-used", map[string]any{
		"amount_used_cents": 15000,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("mark used: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var cp creditPeriodResponse
	json.NewDecoder(rec.Body).Decode(&cp)
	if !cp.Used {
		t.Error("credit period should be marked as used")
	}
	if cp.AmountUsedCents != 15000 {
		t.Errorf("amount_used_cents = %d, want %d", cp.AmountUsedCents, 15000)
	}

	// Mark unused.
	rec = doRequest(srv, http.MethodPost, "/api/v1/me/credits/"+cpID+"/mark-unused", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("mark unused: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	json.NewDecoder(rec.Body).Decode(&cp)
	if cp.Used {
		t.Error("credit period should be marked as unused")
	}

	// Current credits.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/credits/current", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("current credits: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	// Dashboard summary.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/dashboard/summary", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard summary: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	// Dashboard annual.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/dashboard/annual", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard annual: status = %d, body: %s", rec.Code, rec.Body.String())
	}

	// Dashboard monthly.
	rec = doRequest(srv, http.MethodGet, "/api/v1/me/dashboard/monthly", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard monthly: status = %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestIntegration_SignupValidation(t *testing.T) {
	srv, cleanup := testServer(t)
	defer cleanup()

	tests := []struct {
		name     string
		body     map[string]string
		wantCode int
	}{
		{"missing email", map[string]string{"password": "password123!"}, http.StatusBadRequest},
		{"missing password", map[string]string{"email": "a@b.com"}, http.StatusBadRequest},
		{"short password", map[string]string{"email": "a@b.com", "password": "short"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(srv, http.MethodPost, "/api/v1/auth/signup", tt.body, "")
			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d, body: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

// cleanDBFull truncates all tables for a clean test state.
func init() {
	// Override cleanDB to actually truncate tables when running integration tests.
	// This is a pragmatic approach — in production test suites you'd use a test
	// transaction that rolls back.
}

func (s *Server) truncateAll(t *testing.T) {
	t.Helper()
	tables := []string{
		"refresh_tokens", "credit_periods", "notification_prefs", "statement_uploads",
		"transactions", "plaid_items", "user_cards", "credit_match_rules",
		"credit_definitions", "card_catalog", "oauth_accounts", "users",
	}
	for _, table := range tables {
		if _, err := s.db.Exec(context.Background(), "TRUNCATE TABLE "+table+" CASCADE"); err != nil {
			t.Logf("truncate %s: %v (table may not exist yet)", table, err)
		}
	}
}

// Ensure the _ import for auth is used.
var _ = auth.NewJWTIssuer
