package scheduler

import (
	"testing"
	"time"
)

func TestEffectiveRetention(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		base       float64
		targetDate *time.Time
		now        time.Time
		want       float64
	}{
		{
			name:       "no target date returns base",
			base:       0.9,
			targetDate: nil,
			now:        now,
			want:       0.9,
		},
		{
			name:       "more than 90 days out returns base",
			base:       0.9,
			targetDate: ptr(now.AddDate(0, 0, 120)),
			now:        now,
			want:       0.9,
		},
		{
			name:       "exactly 90 days out returns base",
			base:       0.9,
			targetDate: ptr(now.AddDate(0, 0, 90)),
			now:        now,
			want:       0.9,
		},
		{
			name:       "on target date returns ceiling",
			base:       0.9,
			targetDate: ptr(now),
			now:        now,
			want:       0.95,
		},
		{
			name:       "past target date returns ceiling",
			base:       0.9,
			targetDate: ptr(now.AddDate(0, 0, -5)),
			now:        now,
			want:       0.95,
		},
		{
			name:       "halfway through the ramp interpolates",
			base:       0.9,
			targetDate: ptr(now.AddDate(0, 0, 45)),
			now:        now,
			want:       0.925,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveRetention(tt.base, tt.targetDate, tt.now)
			if diff := got - tt.want; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("EffectiveRetention() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr(t time.Time) *time.Time {
	return &t
}
