package share

import (
	"strings"
	"testing"
)

func TestTextCleanSolve(t *testing.T) {
	got := Text(Result{
		Variant: "FreeCell", Day: 200, Moves: 1, Undos: 0, Hints: 0,
		Streak: 1, MaxStreak: 9, URL: "https://example/?v=freecell&d=200",
	})
	// singular "move", clean-solve flair, and the URL on its own last line.
	for _, want := range []string{"Daily #200", "solved in 1 move ", "clean solve", "1-day streak (best 9)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if lines := strings.Split(got, "\n"); lines[len(lines)-1] != "https://example/?v=freecell&d=200" {
		t.Errorf("URL should be the last line:\n%s", got)
	}
}

func TestTextAssisted(t *testing.T) {
	got := Text(Result{
		Variant: "Spider · 4 suits", Day: 3, Moves: 220, Undos: 2, Hints: 1,
		Streak: 4, MaxStreak: 4,
	})
	for _, want := range []string{"220 moves", "2 undos", "1 hint"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "clean solve") {
		t.Errorf("assisted solve should not read as clean:\n%s", got)
	}
}
