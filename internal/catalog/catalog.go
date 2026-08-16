// Package catalog exposes the embedded Neetcode 150 roadmap.
package catalog

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed problems.json
var problemsFS embed.FS

// Difficulty is a problem's difficulty tier.
type Difficulty string

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

// Problem is one entry of the embedded roadmap.
type Problem struct {
	Slug       string     `json:"slug"`
	Title      string     `json:"title"`
	Pattern    string     `json:"pattern"`
	Difficulty Difficulty `json:"difficulty"`
	Order      int        `json:"order"`
}

// Catalog holds the roadmap's patterns and problems, in roadmap order.
type Catalog struct {
	problems  []Problem
	patterns  []string
	byPattern map[string][]Problem
}

// Load parses the embedded problem catalog.
func Load() (*Catalog, error) {
	data, err := problemsFS.ReadFile("problems.json")
	if err != nil {
		return nil, fmt.Errorf("reading embedded catalog: %w", err)
	}

	var problems []Problem
	if err := json.Unmarshal(data, &problems); err != nil {
		return nil, fmt.Errorf("parsing embedded catalog: %w", err)
	}

	patterns := make([]string, 0)
	byPattern := make(map[string][]Problem)
	seen := make(map[string]bool)
	for _, p := range problems {
		if !seen[p.Pattern] {
			seen[p.Pattern] = true
			patterns = append(patterns, p.Pattern)
		}
		byPattern[p.Pattern] = append(byPattern[p.Pattern], p)
	}

	return &Catalog{problems: problems, patterns: patterns, byPattern: byPattern}, nil
}

// Problems returns every problem, in roadmap order.
func (c *Catalog) Problems() []Problem {
	return c.problems
}

// Patterns returns every pattern name, in roadmap order.
func (c *Catalog) Patterns() []string {
	return c.patterns
}

// ByPattern returns the problems belonging to pattern, in roadmap order.
func (c *Catalog) ByPattern(pattern string) []Problem {
	return c.byPattern[pattern]
}
