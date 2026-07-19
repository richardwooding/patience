// Package stats persists tiny per-variant records: games dealt, games won,
// and best move count. Storage is localStorage in the browser and a JSON
// file under the user config dir natively (build-tagged, chipdeck pattern).
package stats

import (
	"encoding/json"
	"sync"
)

// Entry is one variant's record.
type Entry struct {
	Played    int `json:"played"`
	Won       int `json:"won"`
	BestMoves int `json:"bestMoves"`
}

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

var (
	mu     sync.Mutex
	loaded bool
	data   map[string]Entry
)

func ensure() {
	if loaded {
		return
	}
	loaded = true
	data = map[string]Entry{}
	if raw := load(); raw != nil {
		_ = json.Unmarshal(raw, &data)
	}
}

// Get returns the record for a variant ID.
func Get(id string) Entry {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	return data[id]
}

// Record applies an event to a variant's record and persists.
func Record(id string, ev Event) {
	mu.Lock()
	defer mu.Unlock()
	ensure()
	e := data[id]
	ev(&e)
	data[id] = e
	if raw, err := json.Marshal(data); err == nil {
		store(raw)
	}
}
