package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/robin-vidal/neetrack/internal/auth"
)

type upsertLinkRequest struct {
	NeetcodeSlug string `json:"neetcode_slug"`
}

func (s *Server) handleUpsertLink(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	slug := r.PathValue("slug")

	if _, ok := s.catalog.BySlug(slug); !ok {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error:   "unknown_problem",
			Message: "Problème inconnu.",
		})
		return
	}

	var body upsertLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Requête invalide.",
		})
		return
	}

	neetcodeSlug := strings.TrimSpace(body.NeetcodeSlug)
	if neetcodeSlug == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Le slug NeetCode ne peut pas être vide.",
		})
		return
	}

	if err := s.store.UpsertProblemLink(r.Context(), user.ID, slug, neetcodeSlug); err != nil {
		s.logger.ErrorContext(r.Context(), "upserting problem link", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
