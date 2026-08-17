// Package httpapi wires HTTP handlers onto a Server carrying its dependencies.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/store"
)

// Store is what handlers need from the persistence layer.
type Store interface {
	Ping(ctx context.Context) error
	PatternCards(ctx context.Context, userID int64) ([]scheduler.Card, error)
	AttemptIndex(ctx context.Context, userID int64) (scheduler.AttemptIndex, error)
	DoneCounts(ctx context.Context, userID int64) (map[string]int, error)
	RecordAttempt(ctx context.Context, p store.AttemptParams) (store.AttemptResult, error)
	ActivityCounts(ctx context.Context, userID int64, since time.Time) ([]store.DayCount, error)
	Settings(ctx context.Context, userID int64) (store.UserSettings, error)
	UpdateSettings(ctx context.Context, userID int64, settings store.UserSettings) error
	UpsertProblemLink(ctx context.Context, userID int64, problemSlug, neetcodeSlug string) error
	ProblemLinks(ctx context.Context, userID int64) (map[string]string, error)
}

// Server holds the dependencies handlers need, passed in explicitly by main.
type Server struct {
	logger       *slog.Logger
	store        Store
	auth         *auth.Service
	catalog      *catalog.Catalog
	loginLimiter *ipRateLimiter
}

// New builds a Server ready to have its Routes mounted.
func New(logger *slog.Logger, dataStore Store, authService *auth.Service, cat *catalog.Catalog) *Server {
	return &Server{
		logger:       logger,
		store:        dataStore,
		auth:         authService,
		catalog:      cat,
		loginLimiter: newIPRateLimiter(),
	}
}

// Routes returns the HTTP router with every handler mounted.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.Handle("GET /api/me", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleMe)))
	mux.Handle("GET /api/catalog", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleCatalog)))
	mux.Handle("POST /api/sessions", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleComposeSession)))
	mux.Handle("POST /api/attempts", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleCreateAttempt)))
	mux.Handle("GET /api/activity", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleActivity)))
	mux.Handle("GET /api/settings", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleGetSettings)))
	mux.Handle("PATCH /api/settings", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handlePatchSettings)))
	mux.Handle("PUT /api/links/{slug}", auth.RequireAuth(s.auth)(http.HandlerFunc(s.handleUpsertLink)))
	mux.Handle("GET /", spaHandler())
	return mux
}
