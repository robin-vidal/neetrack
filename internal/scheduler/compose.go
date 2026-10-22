package scheduler

import (
	"math"
	"sort"
	"time"

	"github.com/robin-vidal/neetrack/internal/catalog"
)

// Reason says why an Item was picked for a session.
type Reason string

const (
	ReasonReview Reason = "review"
	ReasonNew    Reason = "new"
)

// Item is one problem picked for a session.
type Item struct {
	Problem        catalog.Problem
	Reason         Reason
	Retrievability float64 // set when Reason is ReasonReview
}

// Request configures session composition.
type Request struct {
	Size        int
	NewPct      int
	BlockSize   int
	OpenMax     int
	ExcludeHard bool
}

// AttemptIndex maps a problem slug to the time of its most recent attempt.
type AttemptIndex map[string]time.Time

// Compose builds a session from due patterns and unseen problems.
//
//  1. Patterns due (Due <= now) are sorted by retrievability ascending, most
//     degraded first.
//  2. A review quota is reserved: round(Size * (1 - NewPct/100)), or 0 if
//     nothing is due.
//  3. Each due pattern contributes one problem, up to quota: an unattempted
//     problem of that pattern if one exists, else its least-recently
//     attempted problem.
//  4. The rest is filled with never-attempted problems in roadmap order,
//     capped at BlockSize per pattern with no card yet, 1 per pattern with a
//     card, and OpenMax patterns newly opened this session.
//  5. Any remaining slots are filled with further reviews, cycling through
//     the due patterns again.
//  6. Reviews are sorted before new problems.
//
// When ExcludeHard is set, hard problems are never picked, for review or new.
func Compose(cat *catalog.Catalog, cards []Card, attempts AttemptIndex, req Request, now time.Time) []Item {
	cardByPattern := make(map[string]Card, len(cards))
	for _, c := range cards {
		cardByPattern[c.Pattern] = c
	}

	due := dueByRetrievability(cards, now)

	used := make(map[string]bool)
	if req.ExcludeHard {
		markHardUsed(cat, used)
	}
	var items []Item

	reviewQuota := 0
	if len(due) > 0 {
		reviewQuota = int(math.Round(float64(req.Size) * (1 - float64(req.NewPct)/100)))
	}

	for _, d := range due {
		if len(items) >= reviewQuota {
			break
		}
		if item, ok := pickReview(cat, d, attempts, used); ok {
			items = append(items, item)
		}
	}

	items = append(items, pickNew(cat, cardByPattern, attempts, used, req, len(items))...)

	for _, d := range due {
		if len(items) >= req.Size {
			break
		}
		if item, ok := pickReview(cat, d, attempts, used); ok {
			items = append(items, item)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Reason == ReasonReview && items[j].Reason != ReasonReview
	})

	return items
}

func markHardUsed(cat *catalog.Catalog, used map[string]bool) {
	for _, p := range cat.Problems() {
		if p.Difficulty == catalog.Hard {
			used[p.Slug] = true
		}
	}
}

type duePattern struct {
	pattern        string
	retrievability float64
}

func dueByRetrievability(cards []Card, now time.Time) []duePattern {
	var due []duePattern
	for _, c := range cards {
		if c.Due.After(now) {
			continue
		}
		due = append(due, duePattern{pattern: c.Pattern, retrievability: Retrievability(c, now)})
	}
	sort.Slice(due, func(i, j int) bool { return due[i].retrievability < due[j].retrievability })
	return due
}

func pickReview(cat *catalog.Catalog, d duePattern, attempts AttemptIndex, used map[string]bool) (Item, bool) {
	p := pickProblem(cat.ByPattern(d.pattern), attempts, used)
	if p == nil {
		return Item{}, false
	}
	used[p.Slug] = true
	return Item{Problem: *p, Reason: ReasonReview, Retrievability: d.retrievability}, true
}

// pickProblem returns an unattempted problem if one is available, else the
// least-recently attempted one.
func pickProblem(pool []catalog.Problem, attempts AttemptIndex, used map[string]bool) *catalog.Problem {
	var leastRecent *catalog.Problem
	var leastRecentAt time.Time

	for i := range pool {
		p := &pool[i]
		if used[p.Slug] {
			continue
		}
		at, attempted := attempts[p.Slug]
		if !attempted {
			return p
		}
		if leastRecent == nil || at.Before(leastRecentAt) {
			leastRecent = p
			leastRecentAt = at
		}
	}

	return leastRecent
}

func pickNew(
	cat *catalog.Catalog,
	cardByPattern map[string]Card,
	attempts AttemptIndex,
	used map[string]bool,
	req Request,
	alreadyPicked int,
) []Item {
	var items []Item
	perPattern := make(map[string]int)
	opened := 0

	for _, p := range cat.Problems() {
		if alreadyPicked+len(items) >= req.Size {
			break
		}
		if used[p.Slug] {
			continue
		}
		if _, attempted := attempts[p.Slug]; attempted {
			continue
		}

		_, known := cardByPattern[p.Pattern]
		limit := req.BlockSize
		if known {
			limit = 1
		}

		taken := perPattern[p.Pattern]
		if taken >= limit {
			continue
		}
		if !known && taken == 0 {
			if opened >= req.OpenMax {
				continue
			}
			opened++
		}

		perPattern[p.Pattern] = taken + 1
		used[p.Slug] = true
		items = append(items, Item{Problem: p, Reason: ReasonNew})
	}

	return items
}
