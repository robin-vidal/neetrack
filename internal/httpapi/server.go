// Package httpapi wires HTTP handlers onto a Server carrying its dependencies.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
)

// Pinger reports whether a dependency the server relies on is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server holds the dependencies handlers need, passed in explicitly by main.
type Server struct {
	logger *slog.Logger
	db     Pinger
}

// New builds a Server ready to have its Routes mounted.
func New(logger *slog.Logger, db Pinger) *Server {
	return &Server{logger: logger, db: db}
}

// Routes returns the HTTP router with every handler mounted.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	return mux
}
