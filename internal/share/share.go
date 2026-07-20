// Package share builds the copyable result blurb for a solved daily deal.
// It is pure (no rendering, no I/O) so the text is unit-tested headlessly.
// Emoji are fine here — this string lands in a real text field, not the
// bitmap-font canvas.
package share

import (
	"fmt"
	"strings"
)

// Result is everything the share blurb needs about a solved daily.
type Result struct {
	Variant   string // display name, e.g. "Klondike · draw 1"
	Day       int    // daily number
	Moves     int
	Undos     int
	Hints     int
	Streak    int
	MaxStreak int
	URL       string // deep link to replay this exact daily
}

// Text renders the multi-line share blurb.
func Text(r Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🃏 patience · %s · Daily #%d\n", r.Variant, r.Day)
	fmt.Fprintf(&b, "✅ solved in %d %s · %s\n", r.Moves, plural(r.Moves, "move"), cleanliness(r))
	fmt.Fprintf(&b, "🔥 %d-day streak (best %d)\n", r.Streak, r.MaxStreak)
	b.WriteString(r.URL)
	return b.String()
}

// cleanliness flags an assisted-free solve, else lists the assists used.
func cleanliness(r Result) string {
	if r.Undos == 0 && r.Hints == 0 {
		return "clean solve ✨"
	}
	parts := make([]string, 0, 2)
	if r.Undos > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", r.Undos, plural(r.Undos, "undo")))
	}
	if r.Hints > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", r.Hints, plural(r.Hints, "hint")))
	}
	return strings.Join(parts, " · ")
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
