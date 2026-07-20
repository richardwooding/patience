package solitaire

import "testing"

func TestPyramidDealShape(t *testing.T) {
	g := New(NewPyramid(), 9)
	for i := range 28 {
		if n := len(g.Piles[pyramidBase+i].Cards); n != 1 {
			t.Errorf("slot %d has %d cards, want 1", i, n)
		}
	}
	if n := len(g.Piles[0].Cards); n != 24 { // 52 - 28
		t.Errorf("stock has %d, want 24", n)
	}
}

func TestPyramidExposure(t *testing.T) {
	p := NewPyramid()
	g := New(p, 9)
	// Apex (local slot 0 = pile 3) is covered; bottom row (piles 24-30) is not.
	if p.exposed(g, pyramidBase) {
		t.Error("apex should not be exposed at deal")
	}
	if !p.exposed(g, pyramidBase+21) { // first bottom-row slot
		t.Error("bottom-row slots should be exposed at deal")
	}
}

func TestPyramidPairAndExpose(t *testing.T) {
	p := NewPyramid()
	g := New(p, 9)
	foundation := g.pilesOf(Foundation)[0]

	// Row-5 slot local 15 (pile 18) is covered by bottom-row piles 24 and 25.
	setPile(g, 24, up(Spades, 6))
	setPile(g, 25, up(Hearts, 7)) // 6 + 7 = 13
	before := len(g.Piles[foundation].Cards)

	if err := g.Apply(Move{24, 0, 25}); err != nil { // pair the two children
		t.Fatalf("pairing move: %v", err)
	}
	if len(g.Piles[24].Cards) != 0 || len(g.Piles[25].Cards) != 0 {
		t.Errorf("both paired slots should be empty: %d, %d", len(g.Piles[24].Cards), len(g.Piles[25].Cards))
	}
	if len(g.Piles[foundation].Cards) != before+2 {
		t.Errorf("foundation should gain the pair, got +%d", len(g.Piles[foundation].Cards)-before)
	}
	if !p.exposed(g, 18) {
		t.Error("clearing both children should expose the parent slot")
	}
}

func TestPyramidKingAndPairRules(t *testing.T) {
	p := NewPyramid()
	g := New(p, 9)
	foundation := g.pilesOf(Foundation)[0]

	// A lone exposed King clears to the foundation.
	setPile(g, 26, up(Spades, King))
	if !g.Legal(Move{26, 0, foundation}) {
		t.Error("exposed King should play to the foundation")
	}
	// A non-King single card cannot go to the foundation.
	setPile(g, 27, up(Hearts, 5))
	if g.Legal(Move{27, 0, foundation}) {
		t.Error("a non-King must not go to the foundation")
	}
	// A pair that does not sum to 13 is illegal.
	setPile(g, 28, up(Spades, 5))
	setPile(g, 29, up(Hearts, 9)) // 5 + 9 = 14
	if g.Legal(Move{28, 0, 29}) {
		t.Error("5 + 9 must not pair (sum 14)")
	}
	// A covered slot cannot be the target even with a valid sum.
	setPile(g, 3, up(Spades, 4)) // apex, still covered
	setPile(g, 24, up(Hearts, 9))
	if g.Legal(Move{24, 0, 3}) {
		t.Error("a covered slot must not accept a pair")
	}
}

func TestPyramidStockRecycle(t *testing.T) {
	g := New(NewPyramid(), 9)
	stock0 := len(g.Piles[0].Cards)
	for range stock0 {
		if err := g.TapStock(); err != nil {
			t.Fatalf("draw: %v", err)
		}
	}
	if len(g.Piles[1].Cards) != stock0 {
		t.Fatalf("waste should hold all %d drawn, got %d", stock0, len(g.Piles[1].Cards))
	}
	if err := g.TapStock(); err != nil { // recycle
		t.Fatalf("recycle: %v", err)
	}
	if len(g.Piles[0].Cards) != stock0 || len(g.Piles[1].Cards) != 0 || g.Recycles != 1 {
		t.Fatalf("after recycle: stock %d waste %d recycles %d", len(g.Piles[0].Cards), len(g.Piles[1].Cards), g.Recycles)
	}
}
