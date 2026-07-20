package solitaire

import (
	"testing"
	"time"
)

func TestDayNumber(t *testing.T) {
	cases := []struct {
		date time.Time
		want int
	}{
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(2026, 1, 1, 23, 59, 0, 0, time.UTC), 0}, // same local day
		{time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2026, 7, 20, 14, 0, 0, 0, time.UTC), 200},
		{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), 365}, // 2026 is not a leap year
	}
	for _, c := range cases {
		if got := DayNumber(c.date); got != c.want {
			t.Errorf("DayNumber(%s) = %d, want %d", c.date.Format("2006-01-02"), got, c.want)
		}
	}
}

// TestDailySeedGoldens freezes the seed derivation. These values are shared by
// every player and baked into shared links — if this test fails after a code
// change, the change is breaking, not the test.
func TestDailySeedGoldens(t *testing.T) {
	type key struct {
		id  VariantID
		day int
	}
	goldens := map[key]uint64{
		{Klondike1, 0}:   0x096d3e476161c7fa,
		{Klondike1, 200}: 0xcb45a36df638a66d,
		{Klondike3, 0}:   0xef84b2ae4a2295f8,
		{FreeCellV, 0}:   0x67b7af484cb40e48,
		{Spider1, 0}:     0x9c059effb611b30a,
		{Spider2, 0}:     0x463d6c37a297dde2,
		{Spider4, 200}:   0xe80b6b9940707a50,
	}
	for k, want := range goldens {
		if got := DailySeed(k.id, k.day); got != want {
			t.Errorf("DailySeed(%s, %d) = %#016x, want %#016x", k.id, k.day, got, want)
		}
	}
}

func TestDailySeedDistinct(t *testing.T) {
	// Different variants on the same day, and the same variant across days,
	// must not collide — otherwise "everyone's daily" would repeat.
	seen := map[uint64]string{}
	for _, id := range []VariantID{Klondike1, Klondike3, FreeCellV, Spider1, Spider2, Spider4} {
		for day := 0; day < 400; day++ {
			s := DailySeed(id, day)
			label := string(id) + "/" + time.Duration(day).String()
			if prev, ok := seen[s]; ok {
				t.Fatalf("seed collision: %s and %s both %#x", prev, label, s)
			}
			seen[s] = label
		}
	}
}
