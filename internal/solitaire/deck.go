package solitaire

import "math/rand/v2"

// DeckSpec describes what cards a variant plays with. Klondike/FreeCell use
// one full deck; Spider always deals 104 cards — two full decks at 4 suits,
// four ♠♥ decks at 2 suits, eight ♠ decks at 1 suit.
type DeckSpec struct {
	Decks int
	Suits []Suit
}

// Cards builds the unshuffled cards for the spec, face-down.
func (d DeckSpec) Cards() []Card {
	out := make([]Card, 0, d.Decks*len(d.Suits)*13)
	for range d.Decks {
		for _, s := range d.Suits {
			for r := Ace; r <= King; r++ {
				out = append(out, NewCard(s, r))
			}
		}
	}
	return out
}

// Shuffled returns the spec's cards in seeded Fisher-Yates order. The PCG
// source is stable across platforms; the golden tests would catch drift.
func Shuffled(spec DeckSpec, seed uint64) []Card {
	cards := spec.Cards()
	rng := rand.New(rand.NewPCG(seed, seed))
	for i := len(cards) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
	return cards
}
