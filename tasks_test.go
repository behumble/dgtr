package main

import (
	"testing"
	"time"
)

func TestDueOn(t *testing.T) {
	day := "2026-08-03"
	cases := []struct {
		ts   *string
		want bool
	}{
		{nil, false},
		{strPtr("2026-08-03T12:00:00Z"), true},
		{strPtr("2026-08-03T23:59:59Z"), true},
		{strPtr("2026-08-02T23:59:59Z"), false},
		{strPtr("2026-08-04T00:00:00Z"), false},
		{strPtr("2026-08-03"), true}, // bare date (len>=10 path)
	}
	for _, c := range cases {
		if got := dueOn(c.ts, day); got != c.want {
			t.Errorf("dueOn(%v, %s) = %v, want %v", c.ts, day, got, c.want)
		}
	}
}

func TestDueOnOrBefore(t *testing.T) {
	day := "2026-08-03"
	cases := []struct {
		due  string
		want bool
	}{
		{"", false},
		{"2026-08-01", true},
		{"2026-08-03", true},
		{"2026-08-04", false},
		{"2026-08-03T10:00:00Z", true}, // len>=10 path uses first 10 chars
	}
	for _, c := range cases {
		if got := dueOnOrBefore(c.due, day); got != c.want {
			t.Errorf("dueOnOrBefore(%q, %s) = %v, want %v", c.due, day, got, c.want)
		}
	}
}

func TestIsTodayOrFuture(t *testing.T) {
	now := time.Now()
	if !isTodayOrFuture(now) {
		t.Error("today should be today-or-future")
	}
	if !isTodayOrFuture(now.AddDate(0, 0, 1)) {
		t.Error("tomorrow should be today-or-future")
	}
	if isTodayOrFuture(now.AddDate(0, 0, -1)) {
		t.Error("yesterday should NOT be today-or-future")
	}
}

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

func strPtr(s string) *string { return &s }
