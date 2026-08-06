package main

import (
	"testing"
	"time"
)

func TestSplitJoinLines(t *testing.T) {
	in := "a\nb\nc"
	lines := splitLines(in)
	if len(lines) != 3 {
		t.Fatalf("splitLines len = %d, want 3", len(lines))
	}
	out := joinLines(lines)
	if out != in {
		t.Errorf("roundtrip = %q, want %q", out, in)
	}
	// trailing newline variant
	lines2 := splitLines("a\nb\n")
	if len(lines2) != 2 {
		t.Fatalf("trailing newline split len = %d, want 2", len(lines2))
	}
}

func TestIsUpdatedBetween(t *testing.T) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	end := now

	yesterdayNoon := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, now.Location()).Format(time.RFC3339)
	todayNow := now.Format(time.RFC3339Nano)
	twoDaysAgo := time.Date(now.Year(), now.Month(), now.Day()-2, 12, 0, 0, 0, now.Location()).Format(time.RFC3339)

	cases := []struct {
		updated string
		want    bool
	}{
		{"", false},
		{yesterdayNoon, true},
		{todayNow, true},
		{twoDaysAgo, false},
	}
	for _, c := range cases {
		if got := isUpdatedBetween(c.updated, start, end); got != c.want {
			t.Errorf("isUpdatedBetween(%q, %v, %v) = %v, want %v", c.updated, start, end, got, c.want)
		}
	}
}

func strPtr(s string) *string { return &s }



