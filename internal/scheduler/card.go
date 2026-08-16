package scheduler

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Card is a pattern's derived FSRS state.
type Card struct {
	Pattern string
	fsrs.Card
}

// NewCard returns an unreviewed card for pattern.
func NewCard(pattern string) Card {
	return Card{Pattern: pattern, Card: fsrs.NewCard()}
}

// retrievabilityParams is used only to compute retrievability, which does
// not depend on RequestRetention. It is safe to share across users.
var retrievabilityParams = fsrs.NewFSRS(longTermParams(fsrs.DefaultParam().RequestRetention))

// Retrievability returns card's current probability of recall.
func Retrievability(card Card, now time.Time) float64 {
	return retrievabilityParams.GetRetrievability(card.Card, now)
}

// Review applies rating to card at now, targeting requestRetention, and
// returns the resulting card and the interval, in days, until it is next due.
//
// Learning steps are disabled: even a first review yields a day-scale
// interval, since patterns are reviewed at most a few times per session,
// never minutes apart.
func Review(card Card, now time.Time, rating fsrs.Rating, requestRetention float64) (Card, int) {
	f := fsrs.NewFSRS(longTermParams(requestRetention))
	info := f.Next(card.Card, now, rating)

	return Card{Pattern: card.Pattern, Card: info.Card}, int(info.Card.ScheduledDays)
}

func longTermParams(requestRetention float64) fsrs.Parameters {
	p := fsrs.DefaultParam()
	p.RequestRetention = requestRetention
	p.EnableShortTerm = false
	return p
}
