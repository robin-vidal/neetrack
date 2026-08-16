package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
)

const (
	defaultActivityWeeks = 26
	minActivityWeeks     = 1
	maxActivityWeeks     = 104

	dateLayout = "2006-01-02"
)

type activityDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type activityResponse struct {
	Days   []activityDay `json:"days"`
	Streak int           `json:"streak"`
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	weeks := defaultActivityWeeks
	if raw := r.URL.Query().Get("weeks"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			weeks = clamp(parsed, minActivityWeeks, maxActivityWeeks)
		}
	}

	now := time.Now().UTC()
	since := now.AddDate(0, 0, -weeks*7)

	counts, err := s.store.ActivityCounts(ctx, user.ID, since)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading activity counts", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	days := make([]activityDay, len(counts))
	byDate := make(map[string]int, len(counts))
	for i, c := range counts {
		date := c.Date.Format(dateLayout)
		days[i] = activityDay{Date: date, Count: c.Count}
		byDate[date] = c.Count
	}

	writeJSON(w, http.StatusOK, activityResponse{
		Days:   days,
		Streak: streak(byDate, now),
	})
}

// streak counts consecutive active days ending today, or ending yesterday if
// nothing has happened yet today.
func streak(byDate map[string]int, now time.Time) int {
	day := now.Truncate(24 * time.Hour)
	if byDate[day.Format(dateLayout)] == 0 {
		day = day.AddDate(0, 0, -1)
	}

	count := 0
	for byDate[day.Format(dateLayout)] > 0 {
		count++
		day = day.AddDate(0, 0, -1)
	}

	return count
}
