package solitaire

// Pyramid: 28 cards dealt in a seven-row triangle, plus a stock/waste and a
// foundation. A card is exposed once both cards overlapping it below are gone.
// Remove exposed cards in pairs whose ranks sum to 13 (a King is 13 on its
// own). Clear the whole pyramid to win.
//
// The pairing move reuses the ordinary Move machinery: dropping card A onto an
// exposed slot B (or onto B from the waste) leaves B holding two cards, and
// AfterMove sweeps any slot whose pair sums to 13 into the foundation. So undo
// and win detection need no special cases.
type Pyramid struct {
	children [28][2]int // pile indices covering each slot; {-1,-1} on the bottom row
}

// pyramidBase is the pile index of the first pyramid slot (0=stock, 1=waste,
// 2=foundation, 3..30 = the 28 slots in row-major order).
const pyramidBase = 3

// NewPyramid returns the Pyramid rules with slot coverage precomputed.
func NewPyramid() *Pyramid {
	p := &Pyramid{}
	for r := range 7 {
		for c := 0; c <= r; c++ {
			k := r*(r+1)/2 + c
			if r == 6 {
				p.children[k] = [2]int{-1, -1}
				continue
			}
			below := (r+1)*(r+2)/2 + c
			p.children[k] = [2]int{pyramidBase + below, pyramidBase + below + 1}
		}
	}
	return p
}

func (p *Pyramid) ID() VariantID { return PyramidV }
func (p *Pyramid) Name() string  { return "Pyramid" }

func (p *Pyramid) DeckSpec() DeckSpec {
	return DeckSpec{Decks: 1, Suits: []Suit{Spades, Hearts, Diamonds, Clubs}}
}

// Layout: 0=stock, 1=waste, 2=foundation, 3-30 = 28 pyramid slots (row-major).
func (p *Pyramid) Layout() []PileKind {
	kinds := []PileKind{Stock, Waste, Foundation}
	for range 28 {
		kinds = append(kinds, Tableau)
	}
	return kinds
}

func (p *Pyramid) Deal(g *Game, cards []Card) {
	for i := range 28 {
		g.Piles[pyramidBase+i].Cards = []Card{cards[i].WithFaceUp(true)}
	}
	g.Piles[0].Cards = append(g.Piles[0].Cards, cards[28:]...) // stock, face-down
}

// exposed reports whether a pyramid slot is clear to play (both coverers gone).
func (p *Pyramid) exposed(g *Game, pile int) bool {
	k := pile - pyramidBase
	if k < 0 || k >= 28 {
		return false
	}
	ch := p.children[k]
	if ch[0] < 0 {
		return true // bottom row is always exposed
	}
	return len(g.Piles[ch[0]].Cards) == 0 && len(g.Piles[ch[1]].Cards) == 0
}

func (p *Pyramid) CanPickUp(g *Game, src, idx int) bool {
	s := &g.Piles[src]
	switch s.Kind {
	case Waste:
		return idx == len(s.Cards)-1
	case Tableau:
		return idx == 0 && len(s.Cards) == 1 && p.exposed(g, src)
	default:
		return false
	}
}

func (p *Pyramid) CanDrop(g *Game, src, idx, dst int) bool {
	moving := g.Piles[src].Cards[idx]
	d := &g.Piles[dst]
	switch d.Kind {
	case Foundation:
		return moving.Rank() == King // a lone King clears itself
	case Tableau:
		if !p.exposed(g, dst) || len(d.Cards) != 1 {
			return false
		}
		return int(moving.Rank())+int(d.Cards[0].Rank()) == 13
	default:
		return false
	}
}

// AfterMove sweeps any pyramid slot left holding a pair summing to 13 (formed
// by a pairing drop) into the foundation.
func (p *Pyramid) AfterMove(g *Game) {
	fi := g.pilesOf(Foundation)[0]
	for i := range 28 {
		slot := &g.Piles[pyramidBase+i]
		if len(slot.Cards) == 2 && int(slot.Cards[0].Rank())+int(slot.Cards[1].Rank()) == 13 {
			g.Piles[fi].Cards = append(g.Piles[fi].Cards, slot.Cards...)
			slot.Cards = slot.Cards[:0]
		}
	}
}

// TapStock draws one card to the waste, recycling the waste back into the stock
// (unlimited passes) when the stock is empty.
func (p *Pyramid) TapStock(g *Game) error {
	stock, waste := &g.Piles[0], &g.Piles[1]
	if len(stock.Cards) == 0 {
		if len(waste.Cards) == 0 {
			return ErrStockEmpty
		}
		for i := len(waste.Cards) - 1; i >= 0; i-- {
			stock.Cards = append(stock.Cards, waste.Cards[i].WithFaceUp(false))
		}
		waste.Cards = waste.Cards[:0]
		g.Recycles++
		return nil
	}
	c := stock.Cards[len(stock.Cards)-1]
	stock.Cards = stock.Cards[:len(stock.Cards)-1]
	waste.Cards = append(waste.Cards, c.WithFaceUp(true))
	return nil
}

func (p *Pyramid) Won(g *Game) bool {
	for i := range 28 {
		if len(g.Piles[pyramidBase+i].Cards) > 0 {
			return false
		}
	}
	return true
}

func (p *Pyramid) SafeMoves(*Game) []Move       { return nil }
func (p *Pyramid) AutoCompleteReady(*Game) bool { return false }
