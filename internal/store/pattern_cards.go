package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// PatternCards returns every pattern card belonging to userID.
func (s *Store) PatternCards(ctx context.Context, userID int64) ([]scheduler.Card, error) {
	rows, err := s.queries.PatternCardsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying pattern cards: %w", err)
	}

	cards := make([]scheduler.Card, len(rows))
	for i, r := range rows {
		cards[i] = cardFromFields(r.Pattern, r.Stability, r.Difficulty, r.DueAt, r.LastReviewAt, r.Reps, r.Lapses, r.State)
	}

	return cards, nil
}

func upsertPatternCardParams(userID int64, card scheduler.Card) storedb.UpsertPatternCardParams {
	return storedb.UpsertPatternCardParams{
		UserID:       userID,
		Pattern:      card.Pattern,
		Stability:    card.Stability,
		Difficulty:   card.Difficulty,
		DueAt:        pgtype.Timestamptz{Time: card.Due, Valid: true},
		LastReviewAt: pgtype.Timestamptz{Time: card.LastReview, Valid: true},
		Reps:         int32(card.Reps),
		Lapses:       int32(card.Lapses),
		State:        int16(card.State),
	}
}

func cardFromFields(
	pattern string,
	stability, difficulty float64,
	dueAt, lastReviewAt pgtype.Timestamptz,
	reps, lapses int32,
	state int16,
) scheduler.Card {
	return scheduler.Card{
		Pattern: pattern,
		Card: fsrs.Card{
			Due:        dueAt.Time,
			Stability:  stability,
			Difficulty: difficulty,
			Reps:       uint64(reps),
			Lapses:     uint64(lapses),
			State:      fsrs.State(state),
			LastReview: lastReviewAt.Time,
		},
	}
}
