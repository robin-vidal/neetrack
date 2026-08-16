package scheduler

import (
	"sort"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Attempt is one entry of the append-only attempts log, as needed to replay
// FSRS state.
type Attempt struct {
	Pattern   string
	Rating    fsrs.Rating
	CreatedAt time.Time
}

// Settings holds the retention policy used while replaying.
type Settings struct {
	TargetRetention float64
	TargetDate      *time.Time
}

// Replay reconstructs every pattern's card from scratch by replaying
// attempts in chronological order. It is how FSRS parameters are changed
// without losing review history: delete pattern_cards, then Replay.
func Replay(attempts []Attempt, settings Settings) map[string]Card {
	sorted := make([]Attempt, len(attempts))
	copy(sorted, attempts)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CreatedAt.Before(sorted[j].CreatedAt) })

	cards := make(map[string]Card)
	for _, a := range sorted {
		card, ok := cards[a.Pattern]
		if !ok {
			card = NewCard(a.Pattern)
		}

		retention := EffectiveRetention(settings.TargetRetention, settings.TargetDate, a.CreatedAt)
		card, _ = Review(card, a.CreatedAt, a.Rating, retention)
		cards[a.Pattern] = card
	}

	return cards
}
