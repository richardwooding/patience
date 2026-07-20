package stats

import (
	"encoding/json"
	"testing"
)

func TestApplyWinStreak(t *testing.T) {
	var d Daily

	d = applyWin(d, 10, 120) // first ever win
	if d.Streak != 1 || d.MaxStreak != 1 || d.Wins != 1 || d.BestMoves != 120 {
		t.Fatalf("first win: %+v", d)
	}

	d = applyWin(d, 11, 200) // consecutive day → streak grows, best unchanged
	if d.Streak != 2 || d.BestMoves != 120 {
		t.Fatalf("consecutive win: %+v", d)
	}

	d = applyWin(d, 11, 90) // re-win same day → best improves, streak steady
	if d.Streak != 2 || d.Wins != 2 || d.BestMoves != 90 {
		t.Fatalf("same-day re-win: %+v", d)
	}

	d = applyWin(d, 15, 300) // gap of days → streak restarts, max preserved
	if d.Streak != 1 || d.MaxStreak != 2 {
		t.Fatalf("after gap: %+v", d)
	}

	if !d.SolvedToday(15) || d.SolvedToday(16) {
		t.Fatalf("SolvedToday wrong: %+v", d)
	}
}

func TestLegacyStatsMigration(t *testing.T) {
	// A legacy file was a bare map[string]Entry; ensure() must still load it.
	mu.Lock()
	loaded = true
	entries = map[string]Entry{}
	dailies = map[string]Daily{}
	legacy := []byte(`{"klondike-1":{"played":5,"won":2,"bestMoves":140}}`)
	var p persisted
	if err := json.Unmarshal(legacy, &p); err == nil && (p.Variants != nil || p.Daily != nil) {
		t.Fatal("legacy blob should not parse as the new wrapper")
	}
	_ = json.Unmarshal(legacy, &entries)
	mu.Unlock()

	if e := entries["klondike-1"]; e.Played != 5 || e.Won != 2 || e.BestMoves != 140 {
		t.Fatalf("legacy entry not migrated: %+v", e)
	}

	// Round-trips through the new wrapper.
	raw, _ := json.Marshal(persisted{Variants: entries, Daily: dailies})
	var back persisted
	if err := json.Unmarshal(raw, &back); err != nil || back.Variants["klondike-1"].Played != 5 {
		t.Fatalf("new-format round-trip failed: %s", raw)
	}
}
