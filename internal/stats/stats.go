// Package stats summarizes a user's attempts log into practice statistics.
package stats

import (
	"math"
	"sort"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/scheduler"
)

// RankedLimit caps the length of the best and worst problem lists.
const RankedLimit = 5

// Attempt is one entry of the attempts log, as needed to compute statistics.
type Attempt struct {
	ProblemSlug string
	Pattern     string
	Rating      fsrs.Rating
	Duration    time.Duration // zero when the timer never ran
	UsedHint    bool
	SawSolution bool
	CreatedAt   time.Time
}

// Summary is every statistic derived from one user's attempts log.
type Summary struct {
	Overview     Overview
	Ratings      RatingCounts
	ByDifficulty []DifficultyStats
	ByPattern    []PatternStats
	Best         []ProblemStats
	Worst        []ProblemStats
}

// Overview holds the headline numbers across every attempt.
type Overview struct {
	Attempts          int
	ProblemsAttempted int
	ProblemsSolved    int
	TotalTime         time.Duration
	AverageDuration   time.Duration
	SuccessRate       float64
	HintRate          float64
	SolutionRate      float64
	ActiveDays        int
	LongestStreak     int
}

// RatingCounts counts attempts per rating.
type RatingCounts struct {
	Again int
	Hard  int
	Good  int
	Easy  int
}

// DifficultyStats aggregates the attempts on problems of one difficulty.
type DifficultyStats struct {
	Difficulty      catalog.Difficulty
	Target          time.Duration
	Attempts        int
	AverageDuration time.Duration
	SuccessRate     float64
}

// PatternStats aggregates the attempts on problems of one pattern.
type PatternStats struct {
	Pattern           string
	Attempts          int
	ProblemsAttempted int
	AverageRating     float64
	SuccessRate       float64
	AverageDuration   time.Duration
}

// ProblemStats aggregates the attempts on one problem.
type ProblemStats struct {
	Problem         catalog.Problem
	Attempts        int
	AverageRating   float64
	SuccessRate     float64
	AverageDuration time.Duration
	Pace            float64 // AverageDuration over the difficulty's target, 0 when never timed
	LastRating      fsrs.Rating
}

// Summarize computes every statistic from attempts. An attempt counts as a
// success when rated Good or Easy. Durations average only timed attempts.
// Best and Worst rank problems by average rating, then by pace, and never
// share a problem.
func Summarize(cat *catalog.Catalog, attempts []Attempt) Summary {
	sorted := make([]Attempt, len(attempts))
	copy(sorted, attempts)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].CreatedAt.Before(sorted[j].CreatedAt) })

	var all tally
	var ratings RatingCounts
	hints, solutions := 0, 0
	byDifficulty := make(map[catalog.Difficulty]*tally)
	byPattern := make(map[string]*tally)
	patternProblems := make(map[string]map[string]bool)
	byProblem := make(map[string]*tally)
	lastRating := make(map[string]fsrs.Rating)
	days := make(map[string]bool)

	for _, a := range sorted {
		all.add(a)
		ratings.add(a.Rating)
		if a.UsedHint {
			hints++
		}
		if a.SawSolution {
			solutions++
		}
		days[dayKey(a.CreatedAt)] = true

		tallyFor(byPattern, a.Pattern).add(a)
		if patternProblems[a.Pattern] == nil {
			patternProblems[a.Pattern] = make(map[string]bool)
		}
		patternProblems[a.Pattern][a.ProblemSlug] = true

		p, known := cat.BySlug(a.ProblemSlug)
		if !known {
			continue
		}
		tallyFor(byDifficulty, p.Difficulty).add(a)
		tallyFor(byProblem, p.Slug).add(a)
		lastRating[p.Slug] = a.Rating
	}

	problems := problemStats(cat, byProblem, lastRating)
	best, worst := rank(problems)

	return Summary{
		Overview: Overview{
			Attempts:          all.attempts,
			ProblemsAttempted: len(byProblem),
			ProblemsSolved:    countSolved(byProblem),
			TotalTime:         all.timedSum,
			AverageDuration:   all.averageDuration(),
			SuccessRate:       all.successRate(),
			HintRate:          ratio(hints, all.attempts),
			SolutionRate:      ratio(solutions, all.attempts),
			ActiveDays:        len(days),
			LongestStreak:     longestStreak(days),
		},
		Ratings:      ratings,
		ByDifficulty: difficultyStats(byDifficulty),
		ByPattern:    patternStats(cat, byPattern, patternProblems),
		Best:         best,
		Worst:        worst,
	}
}

type tally struct {
	attempts  int
	successes int
	ratingSum int
	timed     int
	timedSum  time.Duration
}

func (t *tally) add(a Attempt) {
	t.attempts++
	t.ratingSum += int(a.Rating)
	if a.Rating >= fsrs.Good {
		t.successes++
	}
	if a.Duration > 0 {
		t.timed++
		t.timedSum += a.Duration
	}
}

func (t *tally) averageRating() float64 {
	return ratio(t.ratingSum, t.attempts)
}

func (t *tally) successRate() float64 {
	return ratio(t.successes, t.attempts)
}

func (t *tally) averageDuration() time.Duration {
	if t.timed == 0 {
		return 0
	}
	return t.timedSum / time.Duration(t.timed)
}

func tallyFor[K comparable](m map[K]*tally, key K) *tally {
	t, ok := m[key]
	if !ok {
		t = &tally{}
		m[key] = t
	}
	return t
}

func (r *RatingCounts) add(rating fsrs.Rating) {
	switch rating {
	case fsrs.Again:
		r.Again++
	case fsrs.Hard:
		r.Hard++
	case fsrs.Good:
		r.Good++
	case fsrs.Easy:
		r.Easy++
	}
}

func ratio(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total)
}

func dayKey(t time.Time) string {
	return t.UTC().Format(time.DateOnly)
}

func longestStreak(days map[string]bool) int {
	longest := 0
	for key := range days {
		day, _ := time.Parse(time.DateOnly, key)
		if days[dayKey(day.AddDate(0, 0, -1))] {
			continue
		}
		length := 1
		for days[dayKey(day.AddDate(0, 0, length))] {
			length++
		}
		longest = max(longest, length)
	}
	return longest
}

func countSolved(byProblem map[string]*tally) int {
	solved := 0
	for _, t := range byProblem {
		if t.successes > 0 {
			solved++
		}
	}
	return solved
}

func difficultyStats(byDifficulty map[catalog.Difficulty]*tally) []DifficultyStats {
	difficulties := []catalog.Difficulty{catalog.Easy, catalog.Medium, catalog.Hard}
	out := make([]DifficultyStats, len(difficulties))
	for i, d := range difficulties {
		t := tallyFor(byDifficulty, d)
		out[i] = DifficultyStats{
			Difficulty:      d,
			Target:          scheduler.TargetFor(d),
			Attempts:        t.attempts,
			AverageDuration: t.averageDuration(),
			SuccessRate:     t.successRate(),
		}
	}
	return out
}

func patternStats(cat *catalog.Catalog, byPattern map[string]*tally, problems map[string]map[string]bool) []PatternStats {
	var out []PatternStats
	for _, pattern := range cat.Patterns() {
		t, ok := byPattern[pattern]
		if !ok {
			continue
		}
		out = append(out, PatternStats{
			Pattern:           pattern,
			Attempts:          t.attempts,
			ProblemsAttempted: len(problems[pattern]),
			AverageRating:     t.averageRating(),
			SuccessRate:       t.successRate(),
			AverageDuration:   t.averageDuration(),
		})
	}
	return out
}

func problemStats(cat *catalog.Catalog, byProblem map[string]*tally, lastRating map[string]fsrs.Rating) []ProblemStats {
	var out []ProblemStats
	for _, p := range cat.Problems() {
		t, ok := byProblem[p.Slug]
		if !ok {
			continue
		}
		avg := t.averageDuration()
		pace := 0.0
		if avg > 0 {
			pace = float64(avg) / float64(scheduler.TargetFor(p.Difficulty))
		}
		out = append(out, ProblemStats{
			Problem:         p,
			Attempts:        t.attempts,
			AverageRating:   t.averageRating(),
			SuccessRate:     t.successRate(),
			AverageDuration: avg,
			Pace:            pace,
			LastRating:      lastRating[p.Slug],
		})
	}
	return out
}

// rank splits problems into the best and the worst, each at most
// RankedLimit long. Worst is ordered worst first.
func rank(problems []ProblemStats) (best, worst []ProblemStats) {
	sorted := make([]ProblemStats, len(problems))
	copy(sorted, problems)
	sort.SliceStable(sorted, func(i, j int) bool { return ranksAbove(sorted[i], sorted[j]) })

	bestCount := min(RankedLimit, (len(sorted)+1)/2)
	worstCount := min(RankedLimit, len(sorted)/2)

	best = sorted[:bestCount]
	for i := len(sorted) - 1; i >= len(sorted)-worstCount; i-- {
		worst = append(worst, sorted[i])
	}
	return best, worst
}

func ranksAbove(a, b ProblemStats) bool {
	if a.AverageRating != b.AverageRating {
		return a.AverageRating > b.AverageRating
	}
	if paceKey(a) != paceKey(b) {
		return paceKey(a) < paceKey(b)
	}
	return a.Problem.Order < b.Problem.Order
}

// paceKey ranks untimed problems after every timed one.
func paceKey(p ProblemStats) float64 {
	if p.Pace == 0 {
		return math.Inf(1)
	}
	return p.Pace
}
