package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/store"
)

type createAttemptRequest struct {
	ProblemSlug     string `json:"problem_slug"`
	DurationSeconds *int   `json:"duration_seconds"`
	UsedHint        bool   `json:"used_hint"`
	SawSolution     bool   `json:"saw_solution"`
	Rating          *int   `json:"rating"`
}

type createAttemptResponse struct {
	Rating       int       `json:"rating"`
	Pattern      string    `json:"pattern"`
	NextDueAt    time.Time `json:"next_due_at"`
	IntervalDays int       `json:"interval_days"`
}

func (s *Server) handleCreateAttempt(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	var body createAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Requête invalide.",
		})
		return
	}

	problem, ok := s.catalog.BySlug(body.ProblemSlug)
	if !ok {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error:   "unknown_problem",
			Message: "Problème inconnu.",
		})
		return
	}

	var duration time.Duration
	if body.DurationSeconds != nil && *body.DurationSeconds > 0 {
		duration = time.Duration(*body.DurationSeconds) * time.Second
	}

	var clientRating *fsrs.Rating
	if body.Rating != nil {
		rating := fsrs.Rating(*body.Rating)
		clientRating = &rating
	}

	settings, err := s.store.Settings(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	now := time.Now()
	retention := scheduler.EffectiveRetention(settings.TargetRetention, settings.TargetDate, now)

	result, err := s.store.RecordAttempt(ctx, store.AttemptParams{
		UserID:           user.ID,
		Problem:          problem,
		Duration:         duration,
		UsedHint:         body.UsedHint,
		SawSolution:      body.SawSolution,
		ClientRating:     clientRating,
		RequestRetention: retention,
		Now:              now,
	})
	if errors.Is(err, scheduler.ErrRatingRequired) || errors.Is(err, scheduler.ErrInvalidRating) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_attempt",
			Message: "Chrono non lancé : indique une note.",
		})
		return
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "recording attempt", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	writeJSON(w, http.StatusOK, createAttemptResponse{
		Rating:       int(result.Rating),
		Pattern:      result.Pattern,
		NextDueAt:    result.NextDueAt,
		IntervalDays: result.IntervalDays,
	})
}
