package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/robin-vidal/neetrack/internal/auth"
)

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type meResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.loginLimiter.Allow(clientIP(r)) {
		writeJSON(w, http.StatusTooManyRequests, errorResponse{
			Error:   "rate_limited",
			Message: "Trop de tentatives. Réessaie dans une minute.",
		})
		return
	}

	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Requête invalide.",
		})
		return
	}

	cookie, err := s.auth.Login(r.Context(), body.Name, body.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Error:   "invalid_credentials",
			Message: "Identifiant ou mot de passe incorrect.",
		})
		return
	}
	if err != nil {
		s.logger.ErrorContext(r.Context(), "login failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		if err := s.auth.Logout(r.Context(), cookie.Value); err != nil {
			s.logger.ErrorContext(r.Context(), "logout failed", "error", err)
		}
	}

	http.SetCookie(w, s.auth.ClearCookie())
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Error:   "unauthenticated",
			Message: "Non authentifié.",
		})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{ID: user.ID, Name: user.Name, IsAdmin: user.IsAdmin})
}
