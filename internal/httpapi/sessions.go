package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/scheduler"
)

type composeSessionRequest struct {
	Size   int `json:"size"`
	NewPct int `json:"new_pct"`
}

type sessionItem struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Pattern        string   `json:"pattern"`
	Difficulty     string   `json:"difficulty"`
	Reason         string   `json:"reason"`
	Retrievability *float64 `json:"retrievability"`
	NeetcodeSlug   string   `json:"neetcode_slug"`
}

func (s *Server) handleComposeSession(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	var body composeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Requête invalide.",
		})
		return
	}

	size := clamp(body.Size, minSessionSize, maxSessionSize)
	newPct := clamp(body.NewPct, 0, 100)

	settings, err := s.store.Settings(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	cards, err := s.store.PatternCards(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading pattern cards", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	attempts, err := s.store.AttemptIndex(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading attempt index", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	links, err := s.store.ProblemLinks(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading problem links", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	req := scheduler.Request{
		Size:        size,
		NewPct:      newPct,
		BlockSize:   settings.BlockSize,
		OpenMax:     settings.OpenMax,
		ExcludeHard: !settings.IncludeHard,
	}
	items := scheduler.Compose(s.catalog, cards, attempts, req, time.Now())

	resp := make([]sessionItem, len(items))
	for i, it := range items {
		neetcodeSlug := it.Problem.NeetcodeSlug
		if override, ok := links[it.Problem.Slug]; ok {
			neetcodeSlug = override
		}

		item := sessionItem{
			Slug:         it.Problem.Slug,
			Title:        it.Problem.Title,
			Pattern:      it.Problem.Pattern,
			Difficulty:   string(it.Problem.Difficulty),
			Reason:       string(it.Reason),
			NeetcodeSlug: neetcodeSlug,
		}
		if it.Reason == scheduler.ReasonReview {
			r := it.Retrievability
			item.Retrievability = &r
		}
		resp[i] = item
	}

	writeJSON(w, http.StatusOK, resp)
}
