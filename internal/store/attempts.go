package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/stats"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// AttemptParams describes one attempt submission.
type AttemptParams struct {
	UserID           int64
	Problem          catalog.Problem
	Duration         time.Duration
	UsedHint         bool
	SawSolution      bool
	ClientRating     *fsrs.Rating
	RequestRetention float64
	Now              time.Time
}

// AttemptResult is what RecordAttempt reports back to the caller.
type AttemptResult struct {
	Rating       fsrs.Rating
	Pattern      string
	NextDueAt    time.Time
	IntervalDays int
}

// RecordAttempt derives a rating, applies FSRS, and writes the attempt and
// the resulting pattern card in a single transaction.
func (s *Store) RecordAttempt(ctx context.Context, p AttemptParams) (AttemptResult, error) {
	outcome := scheduler.Outcome{
		Difficulty:  p.Problem.Difficulty,
		Duration:    p.Duration,
		UsedHint:    p.UsedHint,
		SawSolution: p.SawSolution,
	}

	rating, err := scheduler.DeriveRating(outcome, p.ClientRating)
	if err != nil {
		return AttemptResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AttemptResult{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := s.queries.WithTx(tx)

	card, err := patternCard(ctx, q, p.UserID, p.Problem.Pattern)
	if err != nil {
		return AttemptResult{}, err
	}

	nextCard, intervalDays := scheduler.Review(card, p.Now, rating, p.RequestRetention)

	var durationSeconds pgtype.Int4
	if p.Duration > 0 {
		durationSeconds = pgtype.Int4{Int32: int32(p.Duration.Seconds()), Valid: true}
	}

	err = q.CreateAttempt(ctx, storedb.CreateAttemptParams{
		UserID:          p.UserID,
		ProblemSlug:     p.Problem.Slug,
		Pattern:         p.Problem.Pattern,
		Rating:          int16(rating),
		DurationSeconds: durationSeconds,
		UsedHint:        p.UsedHint,
		SawSolution:     p.SawSolution,
	})
	if err != nil {
		return AttemptResult{}, fmt.Errorf("creating attempt: %w", err)
	}

	if err := q.UpsertPatternCard(ctx, upsertPatternCardParams(p.UserID, nextCard)); err != nil {
		return AttemptResult{}, fmt.Errorf("upserting pattern card: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return AttemptResult{}, fmt.Errorf("committing transaction: %w", err)
	}

	return AttemptResult{
		Rating:       rating,
		Pattern:      p.Problem.Pattern,
		NextDueAt:    nextCard.Due,
		IntervalDays: intervalDays,
	}, nil
}

func patternCard(ctx context.Context, q *storedb.Queries, userID int64, pattern string) (scheduler.Card, error) {
	row, err := q.PatternCardByUserAndPattern(ctx, storedb.PatternCardByUserAndPatternParams{
		UserID:  userID,
		Pattern: pattern,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return scheduler.NewCard(pattern), nil
	}
	if err != nil {
		return scheduler.Card{}, fmt.Errorf("querying pattern card: %w", err)
	}

	return cardFromFields(row.Pattern, row.Stability, row.Difficulty, row.DueAt, row.LastReviewAt, row.Reps, row.Lapses, row.State), nil
}

// LoggedAttempt is an attempt written with an explicit timestamp, as when
// seeding a history.
type LoggedAttempt struct {
	Problem     catalog.Problem
	Rating      fsrs.Rating
	Duration    time.Duration
	UsedHint    bool
	SawSolution bool
	CreatedAt   time.Time
}

// AppendAttempts writes attempts for userID in a single transaction. It
// leaves pattern cards untouched: call ReplayCards afterwards.
func (s *Store) AppendAttempts(ctx context.Context, userID int64, attempts []LoggedAttempt) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := s.queries.WithTx(tx)

	for _, a := range attempts {
		var durationSeconds pgtype.Int4
		if a.Duration > 0 {
			durationSeconds = pgtype.Int4{Int32: int32(a.Duration.Seconds()), Valid: true}
		}

		err := q.CreateAttemptAt(ctx, storedb.CreateAttemptAtParams{
			UserID:          userID,
			ProblemSlug:     a.Problem.Slug,
			Pattern:         a.Problem.Pattern,
			Rating:          int16(a.Rating),
			DurationSeconds: durationSeconds,
			UsedHint:        a.UsedHint,
			SawSolution:     a.SawSolution,
			CreatedAt:       pgtype.Timestamptz{Time: a.CreatedAt, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("creating attempt: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// AttemptIndex returns the time of each problem's most recent attempt.
func (s *Store) AttemptIndex(ctx context.Context, userID int64) (scheduler.AttemptIndex, error) {
	rows, err := s.queries.LastAttemptPerProblem(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying last attempts: %w", err)
	}

	index := make(scheduler.AttemptIndex, len(rows))
	for _, r := range rows {
		index[r.ProblemSlug] = r.LastAttemptAt.Time
	}

	return index, nil
}

// DoneCounts returns, per pattern, the number of distinct problems with at
// least one attempt rated 2 or better.
func (s *Store) DoneCounts(ctx context.Context, userID int64) (map[string]int, error) {
	rows, err := s.queries.DoneCountByPattern(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying done counts: %w", err)
	}

	counts := make(map[string]int, len(rows))
	for _, r := range rows {
		counts[r.Pattern] = int(r.Done)
	}

	return counts, nil
}

// Attempts returns every attempt belonging to userID, for replay.
func (s *Store) Attempts(ctx context.Context, userID int64) ([]scheduler.Attempt, error) {
	rows, err := s.queries.AttemptsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying attempts: %w", err)
	}

	attempts := make([]scheduler.Attempt, len(rows))
	for i, r := range rows {
		attempts[i] = scheduler.Attempt{
			Pattern:   r.Pattern,
			Rating:    fsrs.Rating(r.Rating),
			CreatedAt: r.CreatedAt.Time,
		}
	}

	return attempts, nil
}

// AttemptLog returns every attempt belonging to userID, oldest first, for
// statistics.
func (s *Store) AttemptLog(ctx context.Context, userID int64) ([]stats.Attempt, error) {
	rows, err := s.queries.AttemptLogByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying attempt log: %w", err)
	}

	attempts := make([]stats.Attempt, len(rows))
	for i, r := range rows {
		attempts[i] = stats.Attempt{
			ProblemSlug: r.ProblemSlug,
			Pattern:     r.Pattern,
			Rating:      fsrs.Rating(r.Rating),
			Duration:    time.Duration(r.DurationSeconds.Int32) * time.Second,
			UsedHint:    r.UsedHint,
			SawSolution: r.SawSolution,
			CreatedAt:   r.CreatedAt.Time,
		}
	}

	return attempts, nil
}

// ProblemMastery returns each attempted problem's mastery level, keyed by
// problem slug. Problems never attempted are absent from the map.
func (s *Store) ProblemMastery(ctx context.Context, userID int64) (map[string]int, error) {
	rows, err := s.queries.ProblemAttemptStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying problem attempt stats: %w", err)
	}

	mastery := make(map[string]int, len(rows))
	for _, r := range rows {
		mastery[r.ProblemSlug] = scheduler.Mastery(true, fsrs.Rating(r.LastRating), int(r.Solves))
	}

	return mastery, nil
}

// ReplayCards deletes every pattern card belonging to userID and
// reconstructs them from scratch by replaying the attempts log.
func (s *Store) ReplayCards(ctx context.Context, userID int64, settings scheduler.Settings) error {
	attempts, err := s.Attempts(ctx, userID)
	if err != nil {
		return err
	}

	cards := scheduler.Replay(attempts, settings)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := s.queries.WithTx(tx)

	if err := q.DeletePatternCardsByUser(ctx, userID); err != nil {
		return fmt.Errorf("deleting pattern cards: %w", err)
	}

	for _, card := range cards {
		if err := q.UpsertPatternCard(ctx, upsertPatternCardParams(userID, card)); err != nil {
			return fmt.Errorf("upserting pattern card: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
