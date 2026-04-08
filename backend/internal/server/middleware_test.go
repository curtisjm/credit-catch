package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"creditcatch/backend/internal/auth"
)

func TestRequireAuth_ValidToken(t *testing.T) {
	jwt := auth.NewJWTIssuer("test-secret", 1*time.Hour)
	s := &Server{jwt: jwt}

	token, err := jwt.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	var gotUserID string
	handler := s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != "user-123" {
		t.Errorf("userID = %q, want %q", gotUserID, "user-123")
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	jwt := auth.NewJWTIssuer("test-secret", 1*time.Hour)
	s := &Server{jwt: jwt}

	handler := s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_InvalidHeader(t *testing.T) {
	jwt := auth.NewJWTIssuer("test-secret", 1*time.Hour)
	s := &Server{jwt: jwt}

	handler := s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "token-without-bearer"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
		{"bearer with spaces", "Bearer   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	jwt := auth.NewJWTIssuer("test-secret", -1*time.Second)
	s := &Server{jwt: jwt}

	token, _ := jwt.Issue("user-123")

	handler := s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret-one", 1*time.Hour)
	validator := auth.NewJWTIssuer("secret-two", 1*time.Hour)
	s := &Server{jwt: validator}

	token, _ := issuer.Issue("user-123")

	handler := s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUserIDFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	uid := UserIDFromContext(req.Context())
	if uid != "" {
		t.Errorf("UserIDFromContext() = %q, want empty", uid)
	}
}
