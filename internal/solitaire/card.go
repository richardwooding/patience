// Package solitaire is the pure rules core: cards, piles, moves, undo, and
// the per-variant Rules implementations. It has no rendering dependencies —
// every behavior is testable headlessly, and the UI is a thin driver.
package solitaire

import "fmt"

// Suit of a card. Spades and Clubs are black; Hearts and Diamonds red.
type Suit uint8

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

// Rank of a card, Ace = 1 through King = 13.
type Rank uint8

const (
	Ace   Rank = 1
	Jack  Rank = 11
	Queen Rank = 12
	King  Rank = 13
)

// Card packs rank (bits 0-3), suit (bits 4-5), and face-up (bit 6) into one
// byte — a whole two-deck Spider game is 104 bytes, which is what makes
// snapshot-per-move undo free.
type Card uint8

// NewCard builds a face-down card.
func NewCard(s Suit, r Rank) Card {
	return Card(uint8(r)&0xF | uint8(s)<<4)
}

func (c Card) Rank() Rank    { return Rank(c & 0xF) }
func (c Card) Suit() Suit    { return Suit(c >> 4 & 0x3) }
func (c Card) FaceUp() bool  { return c&0x40 != 0 }
func (c Card) IsRed() bool   { s := c.Suit(); return s == Hearts || s == Diamonds }
func (c Card) IsBlack() bool { return !c.IsRed() }

// WithFaceUp returns the card with its face-up bit set accordingly.
func (c Card) WithFaceUp(up bool) Card {
	if up {
		return c | 0x40
	}
	return c &^ 0x40
}

var suitRunes = [4]rune{'♠', '♥', '♦', '♣'}
var rankNames = [14]string{"?", "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

// RankName is the display label for a rank ("A", "10", "Q"...).
func (c Card) RankName() string {
	r := c.Rank()
	if r < 1 || r > 13 {
		return "?"
	}
	return rankNames[r]
}

// SuitRune is the card's suit glyph.
func (c Card) SuitRune() rune { return suitRunes[c.Suit()] }

// String renders "Q♥" for face-up cards and "(Q♥)" for face-down — used by
// the seeded-deal golden tests.
func (c Card) String() string {
	s := fmt.Sprintf("%s%c", c.RankName(), c.SuitRune())
	if !c.FaceUp() {
		return "(" + s + ")"
	}
	return s
}
