package store

import (
	"context"
	"fmt"

	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// UpsertProblemLink records userID's corrected NeetCode slug for
// problemSlug.
func (s *Store) UpsertProblemLink(ctx context.Context, userID int64, problemSlug, neetcodeSlug string) error {
	err := s.queries.UpsertProblemLink(ctx, storedb.UpsertProblemLinkParams{
		UserID:       userID,
		ProblemSlug:  problemSlug,
		NeetcodeSlug: neetcodeSlug,
	})
	if err != nil {
		return fmt.Errorf("upserting problem link: %w", err)
	}

	return nil
}

// ProblemLinks returns userID's problem slug to corrected NeetCode slug
// overrides.
func (s *Store) ProblemLinks(ctx context.Context, userID int64) (map[string]string, error) {
	rows, err := s.queries.ProblemLinksByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying problem links: %w", err)
	}

	links := make(map[string]string, len(rows))
	for _, r := range rows {
		links[r.ProblemSlug] = r.NeetcodeSlug
	}

	return links, nil
}
