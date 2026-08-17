package scheduler

import (
	"testing"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestMastery(t *testing.T) {
	tests := []struct {
		name       string
		attempted  bool
		lastRating fsrs.Rating
		solves     int
		want       int
	}{
		{name: "never attempted", attempted: false, want: MasteryNeverAttempted},
		{name: "last attempt failed", attempted: true, lastRating: fsrs.Again, solves: 0, want: MasteryNeedsReview},
		{name: "last attempt failed despite prior solves", attempted: true, lastRating: fsrs.Again, solves: 5, want: MasteryNeedsReview},
		{name: "single hard pass", attempted: true, lastRating: fsrs.Hard, solves: 0, want: MasterySeenOnce},
		{name: "single good pass", attempted: true, lastRating: fsrs.Good, solves: 1, want: MasterySeenOnce},
		{name: "two solves, good", attempted: true, lastRating: fsrs.Good, solves: 2, want: MasterySolid},
		{name: "three solves but not easy last", attempted: true, lastRating: fsrs.Good, solves: 3, want: MasterySolid},
		{name: "three solves, easy last", attempted: true, lastRating: fsrs.Easy, solves: 3, want: MasteryMastered},
		{name: "many solves, easy last", attempted: true, lastRating: fsrs.Easy, solves: 10, want: MasteryMastered},
		{name: "two solves but easy with only two", attempted: true, lastRating: fsrs.Easy, solves: 2, want: MasterySolid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Mastery(tt.attempted, tt.lastRating, tt.solves); got != tt.want {
				t.Errorf("Mastery(%v, %v, %d) = %d, want %d", tt.attempted, tt.lastRating, tt.solves, got, tt.want)
			}
		})
	}
}
