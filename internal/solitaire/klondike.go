package solitaire

// Klondike: 7 tableau columns dealt 1..7 with tops face-up, stock/waste with
// draw-1 or draw-3, four foundations. Unlimited recycles.
type Klondike struct {
	drawCount int
}

// NewKlondike returns the rules for draw-1 or draw-3.
func NewKlondike(draw int) *Klondike { return &Klondike{drawCount: draw} }

func (k *Klondike) ID() VariantID {
	if k.drawCount == 3 {
		return Klondike3
	}
	return Klondike1
}

func (k *Klondike) Name() string {
	if k.drawCount == 3 {
		return "Klondike · draw 3"
	}
	return "Klondike · draw 1"
}

func (k *Klondike) DeckSpec() DeckSpec {
	return DeckSpec{Decks: 1, Suits: []Suit{Spades, Hearts, Diamonds, Clubs}}
}

// Layout: 0=stock, 1=waste, 2-5=foundations, 6-12=tableau.
func (k *Klondike) Layout() []PileKind {
	return []PileKind{
		Stock, Waste,
		Foundation, Foundation, Foundation, Foundation,
		Tableau, Tableau, Tableau, Tableau, Tableau, Tableau, Tableau,
	}
}

func (k *Klondike) Deal(g *Game, cards []Card) {
	next := 0
	for col := range 7 {
		pile := &g.Piles[6+col]
		for row := 0; row <= col; row++ {
			c := cards[next]
			next++
			pile.Cards = append(pile.Cards, c.WithFaceUp(row == col))
		}
	}
	g.Piles[0].Cards = append(g.Piles[0].Cards, cards[next:]...)
}

func (k *Klondike) CanPickUp(g *Game, src, idx int) bool {
	p := &g.Piles[src]
	switch p.Kind {
	case Waste:
		return idx == len(p.Cards)-1
	case Foundation:
		return idx == len(p.Cards)-1
	case Tableau:
		return isRun(p.Cards[idx:], descAltColor)
	default:
		return false
	}
}

func (k *Klondike) CanDrop(g *Game, src, idx, dst int) bool {
	moving := g.Piles[src].Cards[idx:]
	d := &g.Piles[dst]
	switch d.Kind {
	case Tableau:
		top, ok := d.Top()
		if !ok {
			return moving[0].Rank() == King
		}
		return top.FaceUp() && descAltColor(top, moving[0])
	case Foundation:
		return len(moving) == 1 && foundationDrop(d, moving[0])
	default:
		return false
	}
}

func (k *Klondike) AfterMove(g *Game) { g.flipExposed() }

// TapStock draws drawCount cards to the waste, or recycles the waste back
// into the stock when the stock is empty.
func (k *Klondike) TapStock(g *Game) error {
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
	n := min(k.drawCount, len(stock.Cards))
	for range n {
		c := stock.Cards[len(stock.Cards)-1]
		stock.Cards = stock.Cards[:len(stock.Cards)-1]
		waste.Cards = append(waste.Cards, c.WithFaceUp(true))
	}
	return nil
}

func (k *Klondike) Won(g *Game) bool {
	for _, fi := range g.pilesOf(Foundation) {
		if len(g.Piles[fi].Cards) != 13 {
			return false
		}
	}
	return true
}

func (k *Klondike) SafeMoves(g *Game) []Move { return safeFoundationSends(g) }

// AutoCompleteReady: everything face-up and the stock exhausted — from here
// the safe-send loop finishes the game unaided.
func (k *Klondike) AutoCompleteReady(g *Game) bool {
	if len(g.Piles[0].Cards) != 0 || len(g.Piles[1].Cards) != 0 {
		return false
	}
	for _, ti := range g.pilesOf(Tableau) {
		for _, c := range g.Piles[ti].Cards {
			if !c.FaceUp() {
				return false
			}
		}
	}
	return !g.Won()
}
