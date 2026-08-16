package scheduler

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestReplayReconstructsCards(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	settings := Settings{TargetRetention: 0.9}

	attempts := []Attempt{
		{Pattern: "Arrays & Hashing", Rating: fsrs.Good, CreatedAt: base},
		{Pattern: "Arrays & Hashing", Rating: fsrs.Good, CreatedAt: base.AddDate(0, 0, 5)},
		{Pattern: "Two Pointers", Rating: fsrs.Easy, CreatedAt: base.AddDate(0, 0, 1)},
	}

	cards := Replay(attempts, settings)

	if len(cards) != 2 {
		t.Fatalf("len(cards) = %d, want 2", len(cards))
	}

	arrays, ok := cards["Arrays & Hashing"]
	if !ok {
		t.Fatalf("missing card for Arrays & Hashing")
	}
	if arrays.Reps != 2 {
		t.Errorf("Arrays & Hashing Reps = %d, want 2", arrays.Reps)
	}

	twoPointers, ok := cards["Two Pointers"]
	if !ok {
		t.Fatalf("missing card for Two Pointers")
	}
	if twoPointers.Reps != 1 {
		t.Errorf("Two Pointers Reps = %d, want 1", twoPointers.Reps)
	}
}

func TestReplayIsOrderIndependentOfInputSlice(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	settings := Settings{TargetRetention: 0.9}

	inOrder := []Attempt{
		{Pattern: "Arrays & Hashing", Rating: fsrs.Good, CreatedAt: base},
		{Pattern: "Arrays & Hashing", Rating: fsrs.Again, CreatedAt: base.AddDate(0, 0, 5)},
	}
	reversed := []Attempt{inOrder[1], inOrder[0]}

	cardsInOrder := Replay(inOrder, settings)
	cardsReversed := Replay(reversed, settings)

	a := cardsInOrder["Arrays & Hashing"]
	b := cardsReversed["Arrays & Hashing"]

	if a.Stability != b.Stability || a.Lapses != b.Lapses {
		t.Errorf("Replay should sort by CreatedAt regardless of input order: inOrder=%+v reversed=%+v", a, b)
	}
}

func TestReplayEmpty(t *testing.T) {
	cards := Replay(nil, Settings{TargetRetention: 0.9})
	if len(cards) != 0 {
		t.Errorf("len(cards) = %d, want 0", len(cards))
	}
}
