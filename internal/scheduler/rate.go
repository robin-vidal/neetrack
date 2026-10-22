package scheduler

import (
	"errors"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
)

// Target durations used to derive a rating from how long an attempt took.
const (
	TargetEasy   = 12 * time.Minute
	TargetMedium = 25 * time.Minute
	TargetHard   = 40 * time.Minute
)

// Outcome is what the client reports about one attempt.
type Outcome struct {
	Difficulty  catalog.Difficulty
	Duration    time.Duration
	UsedHint    bool
	SawSolution bool
}

// ErrRatingRequired is returned when no duration was recorded and the client
// supplied no fallback rating.
var ErrRatingRequired = errors.New("rating required when no duration was recorded")

// ErrInvalidRating is returned when a client-supplied fallback rating is out
// of the fsrs.Again..fsrs.Easy range.
var ErrInvalidRating = errors.New("invalid rating")

// Rate derives a rating from an attempt's outcome. Applied in order: seeing
// the solution always scores Again; using a hint always scores Hard; past
// this, the rating is driven by how o.Duration compares to the target for
// o.Difficulty.
func Rate(o Outcome) fsrs.Rating {
	if o.SawSolution {
		return fsrs.Again
	}
	if o.UsedHint {
		return fsrs.Hard
	}

	target := TargetFor(o.Difficulty)

	switch {
	case o.Duration > 2*target:
		return fsrs.Hard
	case o.Duration > target:
		return fsrs.Good
	case o.Duration <= target/2:
		return fsrs.Easy
	default:
		return fsrs.Good
	}
}

// DeriveRating returns Rate(o) when a duration was recorded, or clientRating
// when the timer never ran and the client supplied a fallback.
func DeriveRating(o Outcome, clientRating *fsrs.Rating) (fsrs.Rating, error) {
	if o.Duration > 0 {
		return Rate(o), nil
	}
	if clientRating == nil {
		return 0, ErrRatingRequired
	}
	if *clientRating < fsrs.Again || *clientRating > fsrs.Easy {
		return 0, ErrInvalidRating
	}
	return *clientRating, nil
}

// TargetFor returns the target duration for a problem of difficulty d.
func TargetFor(d catalog.Difficulty) time.Duration {
	switch d {
	case catalog.Easy:
		return TargetEasy
	case catalog.Hard:
		return TargetHard
	default:
		return TargetMedium
	}
}
