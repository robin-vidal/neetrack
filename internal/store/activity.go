package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// DayCount is the number of attempts recorded on one UTC day.
type DayCount struct {
	Date  time.Time
	Count int
}

// ActivityCounts returns the number of attempts per UTC day since since,
// for userID.
func (s *Store) ActivityCounts(ctx context.Context, userID int64, since time.Time) ([]DayCount, error) {
	rows, err := s.queries.ActivityCounts(ctx, storedb.ActivityCountsParams{
		UserID:    userID,
		CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("querying activity counts: %w", err)
	}

	counts := make([]DayCount, len(rows))
	for i, r := range rows {
		counts[i] = DayCount{Date: r.Day.Time, Count: int(r.Count)}
	}

	return counts, nil
}
