package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"creditcatch/backend/internal/auth"
	"creditcatch/backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *chi.Mux
	db     *pgxpool.Pool
	config *config.Config
	logger *slog.Logger
	jwt    *auth.JWTIssuer
}

func New(db *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     db,
		config: cfg,
		logger: logger,
		jwt:    auth.NewJWTIssuer(cfg.JWTSecret, cfg.JWTExpiry),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://*.fly.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.logRequest)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))

	s.router.Get("/health", s.handleHealth)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Public auth routes.
		r.Post("/auth/signup", s.handleSignup)
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/refresh", s.handleRefresh)
		r.Post("/auth/logout", s.handleLogout)
		r.Post("/auth/oauth", s.handleOAuth)

		// Public card catalog.
		r.Get("/cards", s.handleListCards)
		r.Get("/cards/{card_id}", s.handleGetCard)

		// Authenticated routes.
		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Get("/me/cards", s.handleListUserCards)
			r.Post("/me/cards", s.handleAddUserCard)
			r.Get("/me/cards/{user_card_id}", s.handleGetUserCard)
			r.Patch("/me/cards/{user_card_id}", s.handleUpdateUserCard)
			r.Delete("/me/cards/{user_card_id}", s.handleDeleteUserCard)

			r.Get("/me/credits", s.handleListCredits)
			r.Get("/me/credits/current", s.handleCurrentCredits)
			r.Post("/me/credits/{credit_period_id}/mark-used", s.handleMarkUsed)
			r.Post("/me/credits/{credit_period_id}/mark-unused", s.handleMarkUnused)

			r.Get("/me/dashboard/summary", s.handleDashboardSummary)
			r.Get("/me/dashboard/annual", s.handleDashboardAnnual)
			r.Get("/me/dashboard/monthly", s.handleDashboardMonthly)
		})
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbOK := true
	if err := s.db.Ping(ctx); err != nil {
		s.logger.Error("health check: database ping failed", "error", err)
		dbOK = false
	}

	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]any{
		"status":   http.StatusText(status),
		"database": dbOK,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}
