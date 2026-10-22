package httpapi

import (
	"math"
	"net/http"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/stats"
)

type statsOverview struct {
	Attempts          int     `json:"attempts"`
	ProblemsAttempted int     `json:"problems_attempted"`
	ProblemsSolved    int     `json:"problems_solved"`
	TotalSeconds      int     `json:"total_seconds"`
	AverageSeconds    int     `json:"average_seconds"`
	SuccessRate       float64 `json:"success_rate"`
	HintRate          float64 `json:"hint_rate"`
	SolutionRate      float64 `json:"solution_rate"`
	ActiveDays        int     `json:"active_days"`
	LongestStreak     int     `json:"longest_streak"`
}

type statsRatings struct {
	Again int `json:"again"`
	Hard  int `json:"hard"`
	Good  int `json:"good"`
	Easy  int `json:"easy"`
}

type statsDifficulty struct {
	Difficulty     string  `json:"difficulty"`
	TargetSeconds  int     `json:"target_seconds"`
	Attempts       int     `json:"attempts"`
	AverageSeconds int     `json:"average_seconds"`
	SuccessRate    float64 `json:"success_rate"`
}

type statsPattern struct {
	Pattern           string  `json:"pattern"`
	Attempts          int     `json:"attempts"`
	ProblemsAttempted int     `json:"problems_attempted"`
	AverageRating     float64 `json:"average_rating"`
	SuccessRate       float64 `json:"success_rate"`
	AverageSeconds    int     `json:"average_seconds"`
}

type statsProblem struct {
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Pattern        string  `json:"pattern"`
	Difficulty     string  `json:"difficulty"`
	NeetcodeSlug   string  `json:"neetcode_slug"`
	Attempts       int     `json:"attempts"`
	AverageRating  float64 `json:"average_rating"`
	SuccessRate    float64 `json:"success_rate"`
	AverageSeconds int     `json:"average_seconds"`
	Pace           float64 `json:"pace"`
	LastRating     int     `json:"last_rating"`
}

type statsResponse struct {
	Overview     statsOverview     `json:"overview"`
	Ratings      statsRatings      `json:"ratings"`
	ByDifficulty []statsDifficulty `json:"by_difficulty"`
	ByPattern    []statsPattern    `json:"by_pattern"`
	Best         []statsProblem    `json:"best"`
	Worst        []statsProblem    `json:"worst"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	attempts, err := s.store.AttemptLog(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading attempt log", "error", err)
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

	writeJSON(w, http.StatusOK, toStatsResponse(stats.Summarize(s.catalog, attempts), links))
}

func toStatsResponse(summary stats.Summary, links map[string]string) statsResponse {
	o := summary.Overview
	resp := statsResponse{
		Overview: statsOverview{
			Attempts:          o.Attempts,
			ProblemsAttempted: o.ProblemsAttempted,
			ProblemsSolved:    o.ProblemsSolved,
			TotalSeconds:      seconds(o.TotalTime),
			AverageSeconds:    seconds(o.AverageDuration),
			SuccessRate:       o.SuccessRate,
			HintRate:          o.HintRate,
			SolutionRate:      o.SolutionRate,
			ActiveDays:        o.ActiveDays,
			LongestStreak:     o.LongestStreak,
		},
		Ratings: statsRatings{
			Again: summary.Ratings.Again,
			Hard:  summary.Ratings.Hard,
			Good:  summary.Ratings.Good,
			Easy:  summary.Ratings.Easy,
		},
		ByDifficulty: make([]statsDifficulty, len(summary.ByDifficulty)),
		ByPattern:    make([]statsPattern, len(summary.ByPattern)),
		Best:         toStatsProblems(summary.Best, links),
		Worst:        toStatsProblems(summary.Worst, links),
	}

	for i, d := range summary.ByDifficulty {
		resp.ByDifficulty[i] = statsDifficulty{
			Difficulty:     string(d.Difficulty),
			TargetSeconds:  seconds(d.Target),
			Attempts:       d.Attempts,
			AverageSeconds: seconds(d.AverageDuration),
			SuccessRate:    d.SuccessRate,
		}
	}

	for i, p := range summary.ByPattern {
		resp.ByPattern[i] = statsPattern{
			Pattern:           p.Pattern,
			Attempts:          p.Attempts,
			ProblemsAttempted: p.ProblemsAttempted,
			AverageRating:     p.AverageRating,
			SuccessRate:       p.SuccessRate,
			AverageSeconds:    seconds(p.AverageDuration),
		}
	}

	return resp
}

func toStatsProblems(problems []stats.ProblemStats, links map[string]string) []statsProblem {
	out := make([]statsProblem, len(problems))
	for i, p := range problems {
		neetcodeSlug := p.Problem.NeetcodeSlug
		if override, ok := links[p.Problem.Slug]; ok {
			neetcodeSlug = override
		}
		out[i] = statsProblem{
			Slug:           p.Problem.Slug,
			Title:          p.Problem.Title,
			Pattern:        p.Problem.Pattern,
			Difficulty:     string(p.Problem.Difficulty),
			NeetcodeSlug:   neetcodeSlug,
			Attempts:       p.Attempts,
			AverageRating:  p.AverageRating,
			SuccessRate:    p.SuccessRate,
			AverageSeconds: seconds(p.AverageDuration),
			Pace:           p.Pace,
			LastRating:     int(p.LastRating),
		}
	}
	return out
}

func seconds(d time.Duration) int {
	return int(math.Round(d.Seconds()))
}
