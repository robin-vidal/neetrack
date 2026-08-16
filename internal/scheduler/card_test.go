package scheduler

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestNewCardRetrievability(t *testing.T) {
	card := NewCard("Arrays & Hashing")
	if got := Retrievability(card, time.Now()); got != 0 {
		t.Errorf("Retrievability(new card) = %v, want 0", got)
	}
}

func TestReviewFirstReviewIsDayScale(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	card := NewCard("Arrays & Hashing")

	tests := []fsrs.Rating{fsrs.Again, fsrs.Hard, fsrs.Good, fsrs.Easy}

	for _, rating := range tests {
		t.Run(rating.String(), func(t *testing.T) {
			next, days := Review(card, now, rating, 0.9)

			if next.State != fsrs.Review {
				t.Errorf("State = %v, want %v (no short-term learning steps)", next.State, fsrs.Review)
			}
			if days < 1 {
				t.Errorf("days = %d, want >= 1", days)
			}
			if next.Due.Before(now.Add(23 * time.Hour)) {
				t.Errorf("Due = %v, want at least a day out from %v", next.Due, now)
			}
			if next.Pattern != "Arrays & Hashing" {
				t.Errorf("Pattern = %q, want %q", next.Pattern, "Arrays & Hashing")
			}
		})
	}
}

func TestReviewHigherRetentionShortensInterval(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	card := NewCard("Arrays & Hashing")
	card, _ = Review(card, now, fsrs.Good, 0.9)

	later := now.AddDate(0, 0, 10)

	_, lowRetentionDays := Review(card, later, fsrs.Good, 0.80)
	_, highRetentionDays := Review(card, later, fsrs.Good, 0.97)

	if highRetentionDays >= lowRetentionDays {
		t.Errorf("higher target retention should shorten the interval: high=%d low=%d", highRetentionDays, lowRetentionDays)
	}
}

func TestReviewAgainLapses(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	card := NewCard("Arrays & Hashing")
	card, _ = Review(card, now, fsrs.Good, 0.9)

	later := now.AddDate(0, 0, 10)
	next, _ := Review(card, later, fsrs.Again, 0.9)

	if next.Lapses != 1 {
		t.Errorf("Lapses = %d, want 1", next.Lapses)
	}
}

func TestRetrievabilityDecaysOverTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	card := NewCard("Arrays & Hashing")
	card, _ = Review(card, now, fsrs.Good, 0.9)

	soon := Retrievability(card, now.Add(time.Hour))
	later := Retrievability(card, now.AddDate(0, 0, 30))

	if later >= soon {
		t.Errorf("retrievability should decay over time: soon=%v later=%v", soon, later)
	}
}
