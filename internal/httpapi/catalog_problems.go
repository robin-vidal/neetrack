package httpapi

import (
	"net/http"

	"github.com/robin-vidal/neetrack/internal/auth"
)

type patternProblem struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Difficulty   string `json:"difficulty"`
	NeetcodeSlug string `json:"neetcode_slug"`
	Mastery      int    `json:"mastery"`
}

func (s *Server) handlePatternProblems(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "missing_pattern",
			Message: "Le paramètre pattern est requis.",
		})
		return
	}

	problems := s.catalog.ByPattern(pattern)
	if problems == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error:   "unknown_pattern",
			Message: "Pattern inconnu.",
		})
		return
	}

	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	mastery, err := s.store.ProblemMastery(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading problem mastery", "error", err)
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

	resp := make([]patternProblem, len(problems))
	for i, p := range problems {
		neetcodeSlug := p.Slug
		if override, ok := links[p.Slug]; ok {
			neetcodeSlug = override
		}

		resp[i] = patternProblem{
			Slug:         p.Slug,
			Title:        p.Title,
			Difficulty:   string(p.Difficulty),
			NeetcodeSlug: neetcodeSlug,
			Mastery:      mastery[p.Slug],
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
