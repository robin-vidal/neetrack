package httpapi

import (
	"net/http"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/scheduler"
)

type patternProgress struct {
	Name           string   `json:"name"`
	Total          int      `json:"total"`
	Done           int      `json:"done"`
	Due            bool     `json:"due"`
	Retrievability *float64 `json:"retrievability"`
}

type catalogResponse struct {
	Patterns []patternProgress `json:"patterns"`
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()
	now := time.Now()

	cards, err := s.store.PatternCards(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading pattern cards", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}
	cardByPattern := make(map[string]scheduler.Card, len(cards))
	for _, c := range cards {
		cardByPattern[c.Pattern] = c
	}

	done, err := s.store.DoneCounts(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading done counts", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	patterns := s.catalog.Patterns()
	resp := catalogResponse{Patterns: make([]patternProgress, 0, len(patterns))}
	for _, name := range patterns {
		progress := patternProgress{
			Name:  name,
			Total: len(s.catalog.ByPattern(name)),
			Done:  done[name],
		}

		if card, ok := cardByPattern[name]; ok {
			progress.Due = !card.Due.After(now)
			r := scheduler.Retrievability(card, now)
			progress.Retrievability = &r
		}

		resp.Patterns = append(resp.Patterns, progress)
	}

	writeJSON(w, http.StatusOK, resp)
}
