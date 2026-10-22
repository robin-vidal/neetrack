package stats

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

var day0 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func attempt(slug string, rating fsrs.Rating, minutes int, dayOffset int) Attempt {
	return Attempt{
		ProblemSlug: slug,
		Pattern:     patternOf[slug],
		Rating:      rating,
		Duration:    time.Duration(minutes) * time.Minute,
		CreatedAt:   day0.AddDate(0, 0, dayOffset),
	}
}

var patternOf = map[string]string{
	"contains-duplicate":  "Arrays & Hashing",
	"valid-anagram":       "Arrays & Hashing",
	"two-sum":             "Arrays & Hashing",
	"group-anagrams":      "Arrays & Hashing",
	"valid-palindrome":    "Two Pointers",
	"3sum":                "Two Pointers",
	"trapping-rain-water": "Two Pointers",
}

func TestSummarizeEmpty(t *testing.T) {
	s := Summarize(testCatalog(t), nil)

	if s.Overview != (Overview{}) {
		t.Errorf("Overview = %+v, want zero value", s.Overview)
	}
	if len(s.ByDifficulty) != 3 {
		t.Errorf("len(ByDifficulty) = %d, want 3 even without attempts", len(s.ByDifficulty))
	}
	if len(s.ByPattern) != 0 || len(s.Best) != 0 || len(s.Worst) != 0 {
		t.Errorf("ByPattern/Best/Worst = %v/%v/%v, want empty", s.ByPattern, s.Best, s.Worst)
	}
}

func TestSummarizeOverview(t *testing.T) {
	attempts := []Attempt{
		attempt("contains-duplicate", fsrs.Easy, 4, 0),
		attempt("valid-anagram", fsrs.Good, 10, 0),
		attempt("group-anagrams", fsrs.Again, 30, 1),
		attempt("group-anagrams", fsrs.Good, 20, 3),
	}
	attempts[2].SawSolution = true
	attempts[3].UsedHint = true
	attempts = append(attempts, Attempt{
		ProblemSlug: "two-sum",
		Pattern:     "Arrays & Hashing",
		Rating:      fsrs.Hard,
		CreatedAt:   day0.AddDate(0, 0, 4),
	})

	o := Summarize(testCatalog(t), attempts).Overview

	want := Overview{
		Attempts:          5,
		ProblemsAttempted: 4,
		ProblemsSolved:    3,
		TotalTime:         64 * time.Minute,
		AverageDuration:   16 * time.Minute,
		SuccessRate:       0.6,
		HintRate:          0.2,
		SolutionRate:      0.2,
		ActiveDays:        4,
		LongestStreak:     2,
	}
	if o != want {
		t.Errorf("Overview = %+v\nwant       %+v", o, want)
	}
}

func TestSummarizeRatingCounts(t *testing.T) {
	attempts := []Attempt{
		attempt("contains-duplicate", fsrs.Easy, 4, 0),
		attempt("valid-anagram", fsrs.Easy, 4, 0),
		attempt("two-sum", fsrs.Good, 10, 0),
		attempt("group-anagrams", fsrs.Hard, 60, 0),
		attempt("3sum", fsrs.Again, 30, 0),
	}

	got := Summarize(testCatalog(t), attempts).Ratings

	want := RatingCounts{Again: 1, Hard: 1, Good: 1, Easy: 2}
	if got != want {
		t.Errorf("Ratings = %+v, want %+v", got, want)
	}
}

func TestSummarizeByDifficulty(t *testing.T) {
	attempts := []Attempt{
		attempt("contains-duplicate", fsrs.Easy, 4, 0),
		attempt("valid-anagram", fsrs.Good, 8, 0),
		attempt("trapping-rain-water", fsrs.Hard, 50, 0),
	}

	byDifficulty := Summarize(testCatalog(t), attempts).ByDifficulty

	tests := []struct {
		difficulty catalog.Difficulty
		attempts   int
		average    time.Duration
		target     time.Duration
		success    float64
	}{
		{catalog.Easy, 2, 6 * time.Minute, 12 * time.Minute, 1},
		{catalog.Medium, 0, 0, 25 * time.Minute, 0},
		{catalog.Hard, 1, 50 * time.Minute, 40 * time.Minute, 0},
	}
	for i, tt := range tests {
		got := byDifficulty[i]
		if got.Difficulty != tt.difficulty || got.Attempts != tt.attempts || got.AverageDuration != tt.average ||
			got.Target != tt.target || got.SuccessRate != tt.success {
			t.Errorf("ByDifficulty[%d] = %+v, want %+v", i, got, tt)
		}
	}
}

func TestSummarizeByPatternFollowsRoadmapOrder(t *testing.T) {
	attempts := []Attempt{
		attempt("3sum", fsrs.Good, 20, 0),
		attempt("contains-duplicate", fsrs.Easy, 4, 1),
		attempt("contains-duplicate", fsrs.Good, 8, 2),
		attempt("valid-anagram", fsrs.Again, 12, 2),
	}

	byPattern := Summarize(testCatalog(t), attempts).ByPattern

	if len(byPattern) != 2 {
		t.Fatalf("len(ByPattern) = %d, want 2", len(byPattern))
	}
	arrays := byPattern[0]
	if arrays.Pattern != "Arrays & Hashing" {
		t.Fatalf("ByPattern[0].Pattern = %q, want Arrays & Hashing first", arrays.Pattern)
	}
	if arrays.Attempts != 3 || arrays.ProblemsAttempted != 2 {
		t.Errorf("Arrays & Hashing attempts/problems = %d/%d, want 3/2", arrays.Attempts, arrays.ProblemsAttempted)
	}
	if arrays.AverageRating != 8.0/3 {
		t.Errorf("AverageRating = %v, want %v", arrays.AverageRating, 8.0/3)
	}
	if arrays.AverageDuration != 8*time.Minute {
		t.Errorf("AverageDuration = %v, want 8m", arrays.AverageDuration)
	}
}

func TestSummarizeBestAndWorst(t *testing.T) {
	attempts := []Attempt{
		attempt("contains-duplicate", fsrs.Easy, 3, 0),
		attempt("valid-anagram", fsrs.Easy, 5, 0),
		attempt("two-sum", fsrs.Good, 10, 0),
		attempt("group-anagrams", fsrs.Again, 40, 0),
		attempt("valid-palindrome", fsrs.Hard, 30, 0),
		attempt("3sum", fsrs.Again, 30, 0),
	}

	s := Summarize(testCatalog(t), attempts)

	wantBest := []string{"contains-duplicate", "valid-anagram", "two-sum"}
	wantWorst := []string{"group-anagrams", "3sum", "valid-palindrome"}
	assertSlugs(t, "Best", s.Best, wantBest)
	assertSlugs(t, "Worst", s.Worst, wantWorst)
}

func TestSummarizeRankingNeverOverlaps(t *testing.T) {
	tests := []struct {
		name      string
		attempts  []Attempt
		wantBest  int
		wantWorst int
	}{
		{
			name:      "single problem is only best",
			attempts:  []Attempt{attempt("two-sum", fsrs.Good, 10, 0)},
			wantBest:  1,
			wantWorst: 0,
		},
		{
			name: "three problems split two and one",
			attempts: []Attempt{
				attempt("two-sum", fsrs.Good, 10, 0),
				attempt("valid-anagram", fsrs.Easy, 4, 0),
				attempt("3sum", fsrs.Again, 30, 0),
			},
			wantBest:  2,
			wantWorst: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Summarize(testCatalog(t), tt.attempts)
			if len(s.Best) != tt.wantBest || len(s.Worst) != tt.wantWorst {
				t.Errorf("len(Best)/len(Worst) = %d/%d, want %d/%d", len(s.Best), len(s.Worst), tt.wantBest, tt.wantWorst)
			}
			seen := make(map[string]bool)
			for _, p := range append(s.Best, s.Worst...) {
				if seen[p.Problem.Slug] {
					t.Errorf("problem %q appears in both lists", p.Problem.Slug)
				}
				seen[p.Problem.Slug] = true
			}
		})
	}
}

func TestSummarizeRankingCappedAtLimit(t *testing.T) {
	cat := testCatalog(t)
	var attempts []Attempt
	for i, p := range cat.Problems()[:20] {
		attempts = append(attempts, Attempt{
			ProblemSlug: p.Slug,
			Pattern:     p.Pattern,
			Rating:      fsrs.Rating(i%4 + 1),
			Duration:    10 * time.Minute,
			CreatedAt:   day0,
		})
	}

	s := Summarize(cat, attempts)

	if len(s.Best) != RankedLimit || len(s.Worst) != RankedLimit {
		t.Errorf("len(Best)/len(Worst) = %d/%d, want %d each", len(s.Best), len(s.Worst), RankedLimit)
	}
}

func TestSummarizeProblemPaceAndLastRating(t *testing.T) {
	attempts := []Attempt{
		attempt("group-anagrams", fsrs.Again, 40, 2),
		attempt("group-anagrams", fsrs.Good, 10, 0),
	}

	s := Summarize(testCatalog(t), attempts)

	p := s.Best[0]
	if p.Pace != 1 {
		t.Errorf("Pace = %v, want 1 (25 min average on a 25 min target)", p.Pace)
	}
	if p.LastRating != fsrs.Again {
		t.Errorf("LastRating = %v, want Again (the chronologically last attempt)", p.LastRating)
	}
}

func TestSummarizeUntimedProblemRanksBelowTimed(t *testing.T) {
	attempts := []Attempt{
		{ProblemSlug: "two-sum", Pattern: "Arrays & Hashing", Rating: fsrs.Good, CreatedAt: day0},
		attempt("valid-anagram", fsrs.Good, 11, 0),
	}

	s := Summarize(testCatalog(t), attempts)

	assertSlugs(t, "Best", s.Best, []string{"valid-anagram"})
	assertSlugs(t, "Worst", s.Worst, []string{"two-sum"})
}

func TestSummarizeIgnoresUnknownProblemsInRankings(t *testing.T) {
	attempts := []Attempt{
		{ProblemSlug: "removed-problem", Pattern: "Arrays & Hashing", Rating: fsrs.Easy, Duration: time.Minute, CreatedAt: day0},
	}

	s := Summarize(testCatalog(t), attempts)

	if s.Overview.Attempts != 1 {
		t.Errorf("Overview.Attempts = %d, want 1", s.Overview.Attempts)
	}
	if s.Overview.ProblemsAttempted != 0 || len(s.Best) != 0 {
		t.Errorf("unknown problem leaked into problem stats: %+v", s)
	}
}

func assertSlugs(t *testing.T, name string, got []ProblemStats, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s has %d problems, want %d", name, len(got), len(want))
	}
	for i := range want {
		if got[i].Problem.Slug != want[i] {
			t.Errorf("%s[%d] = %q, want %q", name, i, got[i].Problem.Slug, want[i])
		}
	}
}
