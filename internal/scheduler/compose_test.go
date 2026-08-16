package scheduler

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
)

func testCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error: %v", err)
	}
	return cat
}

// dueCard builds a card whose Stability leaves it clearly overdue without
// zeroing out its retrievability.
func dueCard(pattern string, now time.Time) Card {
	c := NewCard(pattern)
	c.Stability = 5
	c.State = fsrs.Review
	c.LastReview = now.AddDate(0, 0, -30)
	c.Due = now.AddDate(0, 0, -1)
	return c
}

func notDueCard(pattern string, now time.Time) Card {
	c := NewCard(pattern)
	c.Stability = 30
	c.State = fsrs.Review
	c.LastReview = now
	c.Due = now.AddDate(0, 0, 30)
	return c
}

func TestComposeBlankUser(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	items := Compose(cat, nil, nil, Request{Size: 5, NewPct: 40, BlockSize: 4, OpenMax: 1}, now)

	if len(items) != 4 {
		t.Fatalf("len(items) = %d, want 4 (capped by BlockSize, OpenMax=1 blocks a second pattern)", len(items))
	}
	for _, it := range items {
		if it.Reason != ReasonNew {
			t.Errorf("item %q has reason %v, want %v", it.Problem.Slug, it.Reason, ReasonNew)
		}
		if it.Problem.Pattern != "Arrays & Hashing" {
			t.Errorf("item %q has pattern %q, want first roadmap pattern", it.Problem.Slug, it.Problem.Pattern)
		}
	}
}

func TestComposeEverythingDue(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	cards := []Card{
		dueCard("Arrays & Hashing", now),
		dueCard("Two Pointers", now),
	}

	items := Compose(cat, cards, nil, Request{Size: 2, NewPct: 0, BlockSize: 4, OpenMax: 1}, now)

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	for _, it := range items {
		if it.Reason != ReasonReview {
			t.Errorf("item %q has reason %v, want %v", it.Problem.Slug, it.Reason, ReasonReview)
		}
	}
}

func TestComposeNothingDue(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	cards := []Card{
		notDueCard("Arrays & Hashing", now),
	}

	items := Compose(cat, cards, nil, Request{Size: 3, NewPct: 40, BlockSize: 4, OpenMax: 1}, now)

	for _, it := range items {
		if it.Reason != ReasonNew {
			t.Errorf("item %q has reason %v, want %v (nothing is due)", it.Problem.Slug, it.Reason, ReasonNew)
		}
	}
	// Arrays & Hashing already has a card, so it contributes at most 1 new
	// problem. Two Pointers opens as the session's one new pattern and fills
	// the rest of the request, up to BlockSize.
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	fromArraysHashing := 0
	fromTwoPointers := 0
	for _, it := range items {
		switch it.Problem.Pattern {
		case "Arrays & Hashing":
			fromArraysHashing++
		case "Two Pointers":
			fromTwoPointers++
		default:
			t.Errorf("unexpected pattern %q in items", it.Problem.Pattern)
		}
	}
	if fromArraysHashing != 1 {
		t.Errorf("fromArraysHashing = %d, want 1 (already has a card)", fromArraysHashing)
	}
	if fromTwoPointers != 2 {
		t.Errorf("fromTwoPointers = %d, want 2 (newly opened, filling the rest of the request)", fromTwoPointers)
	}
}

func TestComposeNewPctZeroPrefersReviews(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	cards := []Card{
		dueCard("Arrays & Hashing", now),
	}

	items := Compose(cat, cards, nil, Request{Size: 3, NewPct: 0, BlockSize: 4, OpenMax: 1}, now)

	reviewCount := 0
	for _, it := range items {
		if it.Reason == ReasonReview {
			reviewCount++
		}
	}
	if reviewCount == 0 {
		t.Errorf("expected at least one review item with NewPct=0")
	}
}

func TestComposeNewPctHundredStillFillsWithReviewsWhenNewExhausted(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	cards := []Card{
		dueCard("Tries", now),
	}

	attempts := AttemptIndex{}
	for _, p := range cat.Problems() {
		attempts[p.Slug] = now.Add(-time.Hour)
	}

	items := Compose(cat, cards, attempts, Request{Size: 3, NewPct: 100, BlockSize: 4, OpenMax: 1}, now)

	if len(items) == 0 {
		t.Fatalf("expected leftover review items to fill the session when no new problems remain")
	}
	for _, it := range items {
		if it.Reason != ReasonReview {
			t.Errorf("item %q has reason %v, want %v", it.Problem.Slug, it.Reason, ReasonReview)
		}
	}
}

func TestComposeCatalogExhausted(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	attempts := AttemptIndex{}
	for _, p := range cat.Problems() {
		attempts[p.Slug] = now.Add(-time.Hour)
	}

	items := Compose(cat, nil, attempts, Request{Size: 5, NewPct: 40, BlockSize: 4, OpenMax: 1}, now)

	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 (no cards due, no unattempted problems left)", len(items))
	}
}

func TestComposeReviewsSortedBeforeNew(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	cards := []Card{
		dueCard("Arrays & Hashing", now),
	}

	items := Compose(cat, cards, nil, Request{Size: 5, NewPct: 80, BlockSize: 4, OpenMax: 1}, now)

	sawNew := false
	for _, it := range items {
		if it.Reason == ReasonNew {
			sawNew = true
		}
		if sawNew && it.Reason == ReasonReview {
			t.Fatalf("found a review item after a new item: %+v", items)
		}
	}
}

func TestComposeUnattemptedPreferredOverLeastRecent(t *testing.T) {
	cat := testCatalog(t)
	now := time.Now()

	pattern := "Arrays & Hashing"
	cards := []Card{dueCard(pattern, now)}

	// Attempt every problem except one, so pickProblem must return the
	// unattempted one rather than the least-recently attempted.
	problems := cat.ByPattern(pattern)
	attempts := AttemptIndex{}
	var untouched string
	for i, p := range problems {
		if i == 0 {
			untouched = p.Slug
			continue
		}
		attempts[p.Slug] = now.Add(-time.Duration(i) * time.Hour)
	}

	items := Compose(cat, cards, attempts, Request{Size: 1, NewPct: 0, BlockSize: 4, OpenMax: 1}, now)

	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Problem.Slug != untouched {
		t.Errorf("picked %q, want the unattempted problem %q", items[0].Problem.Slug, untouched)
	}
}
