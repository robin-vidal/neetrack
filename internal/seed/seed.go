// Package seed simulates a plausible practice history, to fill a local
// database with data worth looking at.
package seed

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/scheduler"
)

// Attempt is one simulated attempt.
type Attempt struct {
	Problem     catalog.Problem
	Rating      fsrs.Rating
	Duration    time.Duration
	UsedHint    bool
	SawSolution bool
	CreatedAt   time.Time
}

// Options configures Generate.
type Options struct {
	Days            int
	Now             time.Time
	TargetRetention float64
	Rand            *rand.Rand
}

// Generate simulates opts.Days of practice ending before opts.Now. Each
// active day composes a session with the real scheduler, draws an outcome
// per problem from a per-pattern skill that grows with practice, rates it
// with scheduler.Rate and reviews the pattern card, so the history is one
// the app could have produced itself.
func Generate(cat *catalog.Catalog, opts Options) []Attempt {
	sim := simulation{
		cat:      cat,
		rng:      opts.Rand,
		opts:     opts,
		cards:    make(map[string]scheduler.Card),
		index:    make(scheduler.AttemptIndex),
		skill:    make(map[string]float64),
		practice: make(map[string]int),
	}
	for _, pattern := range cat.Patterns() {
		sim.skill[pattern] = 0.15 + 0.45*sim.rng.Float64()
	}

	start := opts.Now.UTC().Truncate(24*time.Hour).AddDate(0, 0, -opts.Days)
	for d := range opts.Days {
		day := start.AddDate(0, 0, d)
		if sim.rng.Float64() < activeProbability(day) {
			sim.practiceDay(day)
		}
	}

	return sim.attempts
}

type simulation struct {
	cat      *catalog.Catalog
	rng      *rand.Rand
	opts     Options
	cards    map[string]scheduler.Card
	index    scheduler.AttemptIndex
	skill    map[string]float64
	practice map[string]int
	attempts []Attempt
}

func activeProbability(day time.Time) float64 {
	if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		return 0.35
	}
	return 0.6
}

func (s *simulation) practiceDay(day time.Time) {
	at := day.Add(8*time.Hour + time.Duration(s.rng.IntN(13*60))*time.Minute)
	req := scheduler.Request{Size: 2 + s.rng.IntN(4), NewPct: 40, BlockSize: 4, OpenMax: 1}

	for _, item := range scheduler.Compose(s.cat, s.cardList(), s.index, req, at) {
		attempt := s.attempt(item.Problem, at)
		if attempt.CreatedAt.After(s.opts.Now) {
			return
		}
		s.record(attempt)
		at = attempt.CreatedAt.Add(time.Duration(2+s.rng.IntN(9)) * time.Minute)
	}
}

func (s *simulation) cardList() []scheduler.Card {
	var cards []scheduler.Card
	for _, pattern := range s.cat.Patterns() {
		if c, ok := s.cards[pattern]; ok {
			cards = append(cards, c)
		}
	}
	return cards
}

func (s *simulation) attempt(p catalog.Problem, startedAt time.Time) Attempt {
	skill := math.Min(0.95, s.skill[p.Pattern]+0.03*float64(s.practice[p.Pattern]))
	struggle := (1 - skill) * difficultyWeight(p.Difficulty)

	sawSolution := s.rng.Float64() < 0.3*struggle
	usedHint := !sawSolution && s.rng.Float64() < 0.35*struggle

	factor := math.Exp(0.35*s.rng.NormFloat64()) * (0.3 + 0.9*struggle)
	if sawSolution {
		factor = 1.2 + s.rng.Float64()
	}
	target := scheduler.TargetFor(p.Difficulty)
	duration := time.Duration(factor * float64(target)).Round(time.Second)
	duration = min(max(duration, 2*time.Minute), 3*target)

	outcome := scheduler.Outcome{
		Difficulty:  p.Difficulty,
		Duration:    duration,
		UsedHint:    usedHint,
		SawSolution: sawSolution,
	}

	return Attempt{
		Problem:     p,
		Rating:      scheduler.Rate(outcome),
		Duration:    duration,
		UsedHint:    usedHint,
		SawSolution: sawSolution,
		CreatedAt:   startedAt.Add(duration),
	}
}

func difficultyWeight(d catalog.Difficulty) float64 {
	switch d {
	case catalog.Easy:
		return 0.6
	case catalog.Hard:
		return 1.7
	default:
		return 1
	}
}

func (s *simulation) record(a Attempt) {
	pattern := a.Problem.Pattern
	card, ok := s.cards[pattern]
	if !ok {
		card = scheduler.NewCard(pattern)
	}
	s.cards[pattern], _ = scheduler.Review(card, a.CreatedAt, a.Rating, s.opts.TargetRetention)
	s.index[a.Problem.Slug] = a.CreatedAt
	s.practice[pattern]++
	s.attempts = append(s.attempts, a)
}
