package seed

import (
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/scheduler"
)

var now = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

func generate(t *testing.T, days int, seed uint64) []Attempt {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error: %v", err)
	}
	return Generate(cat, Options{
		Days:            days,
		Now:             now,
		TargetRetention: 0.9,
		Rand:            rand.New(rand.NewPCG(seed, seed)),
	})
}

func TestGenerateIsDeterministic(t *testing.T) {
	a := generate(t, 60, 7)
	b := generate(t, 60, 7)

	if !reflect.DeepEqual(a, b) {
		t.Errorf("two runs with the same seed produced different histories")
	}
}

func TestGenerateStaysInWindowAndInOrder(t *testing.T) {
	attempts := generate(t, 90, 1)

	if len(attempts) == 0 {
		t.Fatal("Generate produced no attempts over 90 days")
	}
	earliest := now.Truncate(24*time.Hour).AddDate(0, 0, -90)
	for i, a := range attempts {
		if a.CreatedAt.Before(earliest) || a.CreatedAt.After(now) {
			t.Errorf("attempt %d at %v is outside [%v, %v]", i, a.CreatedAt, earliest, now)
		}
		if i > 0 && a.CreatedAt.Before(attempts[i-1].CreatedAt) {
			t.Errorf("attempt %d at %v precedes attempt %d at %v", i, a.CreatedAt, i-1, attempts[i-1].CreatedAt)
		}
	}
}

func TestGenerateRatingsMatchScheduler(t *testing.T) {
	for i, a := range generate(t, 90, 2) {
		want := scheduler.Rate(scheduler.Outcome{
			Difficulty:  a.Problem.Difficulty,
			Duration:    a.Duration,
			UsedHint:    a.UsedHint,
			SawSolution: a.SawSolution,
		})
		if a.Rating != want {
			t.Errorf("attempt %d rated %v, want %v from its outcome", i, a.Rating, want)
		}
		if a.UsedHint && a.SawSolution {
			t.Errorf("attempt %d both used a hint and saw the solution", i)
		}
	}
}

func TestGenerateRevisitsProblems(t *testing.T) {
	seen := make(map[string]bool)
	revisits := 0
	for _, a := range generate(t, 120, 3) {
		if seen[a.Problem.Slug] {
			revisits++
		}
		seen[a.Problem.Slug] = true
	}

	if revisits == 0 {
		t.Error("no problem was attempted twice, reviews never happened")
	}
}
