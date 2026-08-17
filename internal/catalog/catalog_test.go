package catalog

import "testing"

func TestLoad(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	problems := c.Problems()
	if len(problems) != 150 {
		t.Errorf("len(Problems()) = %d, want 150", len(problems))
	}

	patterns := c.Patterns()
	if len(patterns) != 18 {
		t.Errorf("len(Patterns()) = %d, want 18", len(patterns))
	}

	seenSlugs := make(map[string]bool)
	seenNeetcodeSlugs := make(map[string]string)
	var total int
	for _, pattern := range patterns {
		byPattern := c.ByPattern(pattern)
		if len(byPattern) == 0 {
			t.Errorf("ByPattern(%q) is empty", pattern)
		}
		total += len(byPattern)
		for _, p := range byPattern {
			if p.Pattern != pattern {
				t.Errorf("ByPattern(%q) returned problem with pattern %q", pattern, p.Pattern)
			}
			if seenSlugs[p.Slug] {
				t.Errorf("duplicate slug %q", p.Slug)
			}
			seenSlugs[p.Slug] = true
			if p.Slug == "" || p.Title == "" {
				t.Errorf("problem with empty slug or title: %+v", p)
			}
			if p.NeetcodeSlug == "" {
				t.Errorf("problem %q has empty NeetcodeSlug", p.Slug)
			}
			if other, dup := seenNeetcodeSlugs[p.NeetcodeSlug]; dup {
				t.Errorf("neetcode_slug %q used by both %q and %q", p.NeetcodeSlug, other, p.Slug)
			}
			seenNeetcodeSlugs[p.NeetcodeSlug] = p.Slug
			switch p.Difficulty {
			case Easy, Medium, Hard:
			default:
				t.Errorf("problem %q has unexpected difficulty %q", p.Slug, p.Difficulty)
			}
		}
	}

	if total != len(problems) {
		t.Errorf("sum of ByPattern() lengths = %d, want %d", total, len(problems))
	}
}

func TestByPatternUnknown(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := c.ByPattern("does not exist"); got != nil {
		t.Errorf("ByPattern(unknown) = %v, want nil", got)
	}
}

func TestBySlug(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	p, ok := c.BySlug("two-sum")
	if !ok {
		t.Fatalf("BySlug(%q) not found", "two-sum")
	}
	if p.Title != "Two Sum" {
		t.Errorf("BySlug(%q).Title = %q, want %q", "two-sum", p.Title, "Two Sum")
	}
	if p.NeetcodeSlug != "two-integer-sum" {
		t.Errorf("BySlug(%q).NeetcodeSlug = %q, want %q", "two-sum", p.NeetcodeSlug, "two-integer-sum")
	}

	if _, ok := c.BySlug("does-not-exist"); ok {
		t.Errorf("BySlug(unknown) found a problem, want not found")
	}
}
