// Package stats persists tiny per-variant records: casual games dealt/won with
// a best move count, and daily-deal streaks. Storage is localStorage in the
// browser and a JSON file under the user config dir natively (build-tagged,
// chipdeck pattern).
package stats

import (
	"encoding/json"
	"sync"
)

// Entry is one variant's casual record.
type Entry struct {
	Played    int `json:"played"`
	Won       int `json:"won"`
	BestMoves int `json:"bestMoves"`
}

// Daily is one variant's daily-deal record.
type Daily struct {
	LastWinDay int `json:"lastWinDay"` // day number of the most recent win
	Streak     int `json:"streak"`     // consecutive winning days ending at LastWinDay
	MaxStreak  int `json:"maxStreak"`
	Wins       int `json:"wins"`
	BestMoves  int `json:"bestMoves"`
}

// SolvedToday reports whether day's daily has already been won.
func (d Daily) SolvedToday(day int) bool { return d.Wins > 0 && d.LastWinDay == day }

// Event mutates an Entry.
type Event func(*Entry)

// Dealt counts a new game.
func Dealt(e *Entry) { e.Played++ }

// WonIn counts a win with its move count.
func WonIn(moves int) Event {
	return func(e *Entry) {
		e.Won++
		if e.BestMoves == 0 || moves < e.BestMoves {
			e.BestMoves = moves
		}
	}
}

// applyWin returns d advanced for a win on the given day. It is pure (no
// persistence) so the streak rules can be tested directly. Re-winning the same
// day only updates the best move count; a win on the day after LastWinDay
// extends the streak, and any longer gap restarts it at 1.
func applyWin(d Daily, day, moves int) Daily {
	better := func() {
		if d.BestMoves == 0 || moves < d.BestMoves {
			d.BestMoves = moves
		}
	}
	if d.SolvedToday(day) {
		better()
		return d
	}
	if d.Wins > 0 && d.LastWinDay == day-1 {
		d.Streak++
	} else {
		d.Streak = 1
	}
	d.LastWinDay = day
	d.Wins++
	if d.Streak > d.MaxStreak {
		d.MaxStreak = d.Streak
	}
	better()
	return d
}

// persisted is the on-disk shape. Legacy files were a bare map[string]Entry;
// ensure() migrates those.
type persisted struct {
	Variants map[string]Entry `json:"variants"`
	Daily    map[string]Daily `json:"daily"`
}

var (
	mu      sync.Mutex
	loaded  bool
	entries map[string]Entry
	dailies map[string]Daily
)

func ensure() {
	if loaded {
		return
	}
	loaded = true
	entries = map[string]Entry{}
	dailies = map[string]Daily{}
	raw := load()
	if raw == nil {
		return
	}
	var p persisted
	if err := json.Unmarshal(raw, &p); err == nil && (p.Variants != nil || p.Daily != nil) {
		if p.Variants != nil {
			entries = p.Variants
		}
		if p.Daily != nil {
			dailies = p.Daily
		}
		return
	}
	_ = json.Unmarshal(raw, &entries) // legacy: bare map[string]Entry
}

func persist() {
	if raw, err := json.Marshal(persisted{Variants: entries, Daily: dailies}); err == nil {
		store(raw)
	}
}

// Get returns the casual record for a variant ID.
func Get(id string) Entry {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	return entries[id]
}

// Record applies an event to a variant's casual record and persists.
func Record(id string, ev Event) {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	e := entries[id]
	ev(&e)
	entries[id] = e
	persist()
}

// GetDaily returns the daily record for a variant ID.
func GetDaily(id string) Daily {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	return dailies[id]
}

// RecordDailyWin records a daily-deal win and returns the updated record.
func RecordDailyWin(id string, day, moves int) Daily {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	d := applyWin(dailies[id], day, moves)
	dailies[id] = d
	persist()
	return d
}
