package solitaire

import "testing"

func TestGolfDealShape(t *testing.T) {
	g := New(NewGolf(), 3)
	for _, ti := range g.pilesOf(Tableau) {
		if got := len(g.Piles[ti].Cards); got != 5 {
			t.Errorf("column %d has %d cards, want 5", ti, got)
		}
	}
	if n := len(g.Piles[1].Cards); n != 1 {
		t.Errorf("foundation seeded with %d, want 1", n)
	}
	if n := len(g.Piles[0].Cards); n != 16 { // 52 - 35 - 1
		t.Errorf("stock has %d, want 16", n)
	}
}

func TestGolfMoves(t *testing.T) {
	g := New(NewGolf(), 3)
	setPile(g, 1, up(Spades, 7)) // foundation top 7
	setPile(g, 2, up(Hearts, 8)) // 8 is adjacent up
	setPile(g, 3, up(Clubs, 6))  // 6 is adjacent down
	setPile(g, 4, up(Spades, 9)) // 9 is two away

	if !g.Legal(Move{2, 0, 1}) {
		t.Error("8 should play onto 7")
	}
	if !g.Legal(Move{3, 0, 1}) {
		t.Error("6 should play onto 7")
	}
	if g.Legal(Move{4, 0, 1}) {
		t.Error("9 must not play onto 7 (gap of 2)")
	}

	// No wrap: King and Ace are not adjacent.
	setPile(g, 1, up(Spades, 13))
	setPile(g, 5, up(Hearts, 1))
	if g.Legal(Move{5, 0, 1}) {
		t.Error("Ace must not play onto King (no wrap)")
	}
}

func TestGolfStockAndWin(t *testing.T) {
	g := New(NewGolf(), 3)
	// empty every column, leave one card that can't play, then verify a stock
	// tap moves a card to the foundation and clearing all columns wins.
	for _, ti := range g.pilesOf(Tableau) {
		setPile(g, ti)
	}
	if !g.Won() {
		t.Fatal("all columns empty should be a win")
	}

	fresh := New(NewGolf(), 3)
	before := len(fresh.Piles[1].Cards)
	if err := fresh.TapStock(); err != nil {
		t.Fatalf("TapStock: %v", err)
	}
	if len(fresh.Piles[1].Cards) != before+1 {
		t.Error("stock tap should add one card to the foundation")
	}
}
