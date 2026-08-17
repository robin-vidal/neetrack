package scheduler

import "github.com/open-spaced-repetition/go-fsrs/v3"

// Mastery levels for a single problem, derived from its attempt history.
const (
	MasteryNeverAttempted = 0
	MasteryNeedsReview    = 1
	MasterySeenOnce       = 2
	MasterySolid          = 3
	MasteryMastered       = 4
)

// Mastery derives a problem's mastery level from its last rating and how
// many of its attempts scored Good or better. It mirrors the review
// history the same way FSRS scheduling does, but is a display heuristic
// only: it never feeds back into scheduling.
func Mastery(attempted bool, lastRating fsrs.Rating, solves int) int {
	if !attempted {
		return MasteryNeverAttempted
	}
	if lastRating == fsrs.Again {
		return MasteryNeedsReview
	}
	if solves >= 3 && lastRating == fsrs.Easy {
		return MasteryMastered
	}
	if solves >= 2 {
		return MasterySolid
	}
	return MasterySeenOnce
}
