package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RequireAuth rejects requests without a valid session cookie with 401, and
// injects the authenticated User into the request context otherwise.
func RequireAuth(service *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				writeUnauthenticated(w)
				return
			}

			user, err := service.Authenticate(r.Context(), cookie.Value)
			if errors.Is(err, ErrSessionNotFound) {
				writeUnauthenticated(w)
				return
			}
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorBody{
					Error:   "internal_error",
					Message: "Erreur interne, réessaie plus tard.",
				})
				return
			}

			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
		})
	}
}

func writeUnauthenticated(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, errorBody{
		Error:   "unauthenticated",
		Message: "Non authentifié.",
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
