// Package httpapi wires HTTP handlers onto a Server carrying its dependencies.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/catalog"
)

// Pinger reports whether a dependency the server relies on is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// CatalogSource is the embedded problem roadmap.
type CatalogSource interface {
	Patterns() []string
	ByPattern(name string) []catalog.Problem
}

// Server holds the dependencies handlers need, passed in explicitly by main.
type Server struct {
	logger       *slog.Logger
	db           Pinger
	auth         *auth.Service
	catalog      CatalogSource
	loginLimiter *ipRateLimiter
}

// New builds a Server ready to have its Routes mounted.
func New(logger *slog.Logger, db Pinger, authService *auth.Service, catalogSource CatalogSource) *Server {
	return &Server{
		logger:       logger,
		db:           db,
		auth:         authService,
		catalog:      catalogSource,
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
	return mux
}
