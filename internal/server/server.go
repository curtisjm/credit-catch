package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"creditcatch/backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *chi.Mux
	db     *pgxpool.Pool
	config *config.Config
	logger *slog.Logger
}

func New(db *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     db,
		config: cfg,
		logger: logger,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.logRequest)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))

	s.router.Get("/health", s.handleHealth)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Future endpoints registered here.
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   http.StatusText(status),
		"database": dbOK,
	})
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
