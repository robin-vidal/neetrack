package httpapi

import "net/http"

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
	patterns := s.catalog.Patterns()

	resp := catalogResponse{Patterns: make([]patternProgress, 0, len(patterns))}
	for _, name := range patterns {
		resp.Patterns = append(resp.Patterns, patternProgress{
			Name:  name,
			Total: len(s.catalog.ByPattern(name)),
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
