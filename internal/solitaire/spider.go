package solitaire

// Spider: 10 tableau columns (six cards in the first four, five in the rest,
// tops face-up), 50 stock cards dealt a full row at a time, 8 foundations
// that receive completed same-suit K→A runs. Suits: 1, 2, or 4.
type Spider struct {
	suits int
}

// NewSpider returns Spider rules at the given suit count (1, 2, or 4).
func NewSpider(suits int) *Spider { return &Spider{suits: suits} }

func (s *Spider) ID() VariantID {
	switch s.suits {
	case 1:
		return Spider1
	case 2:
		return Spider2
	default:
		return Spider4
	}
}

func (s *Spider) Name() string {
	switch s.suits {
	case 1:
		return "Spider · 1 suit"
	case 2:
		return "Spider · 2 suits"
	default:
		return "Spider · 4 suits"
	}
}

func (s *Spider) DeckSpec() DeckSpec {
	switch s.suits {
	case 1:
		return DeckSpec{Decks: 8, Suits: []Suit{Spades}}
	case 2:
		return DeckSpec{Decks: 4, Suits: []Suit{Spades, Hearts}}
	default:
		return DeckSpec{Decks: 2, Suits: []Suit{Spades, Hearts, Diamonds, Clubs}}
	}
}

// Layout: 0=stock, 1-8=foundations, 9-18=tableau.
func (s *Spider) Layout() []PileKind {
	kinds := []PileKind{Stock}
	for range 8 {
		kinds = append(kinds, Foundation)
	}
	for range 10 {
		kinds = append(kinds, Tableau)
	}
	return kinds
}

func (s *Spider) Deal(g *Game, cards []Card) {
	next := 0
	for col := range 10 {
		count := 5
		if col < 4 {
			count = 6
		}
		pile := &g.Piles[9+col]
		for row := range count {
			c := cards[next]
			next++
			pile.Cards = append(pile.Cards, c.WithFaceUp(row == count-1))
		}
	}
	g.Piles[0].Cards = append(g.Piles[0].Cards, cards[next:]...)
}

func (s *Spider) CanPickUp(g *Game, src, idx int) bool {
	p := &g.Piles[src]
	if p.Kind != Tableau {
		return false
	}
	return isRun(p.Cards[idx:], descSameSuit)
}

func (s *Spider) CanDrop(g *Game, src, idx, dst int) bool {
	d := &g.Piles[dst]
	if d.Kind != Tableau {
		return false
	}
	moving := g.Piles[src].Cards[idx:]
	top, ok := d.Top()
	if !ok {
		return true
	}
	return top.FaceUp() && top.Rank() == moving[0].Rank()+1
}

// AfterMove removes completed same-suit K→A runs (looping to a fixpoint —
// a removal can expose a card that completes another) and flips exposed
// face-down tops. Removal is an ordinary pile move onto a foundation, so
// undo and win detection need no special cases.
func (s *Spider) AfterMove(g *Game) {
	for {
		g.flipExposed()
		if !s.removeOneRun(g) {
			return
		}
	}
}

func (s *Spider) removeOneRun(g *Game) bool {
	for _, ti := range g.pilesOf(Tableau) {
		p := &g.Piles[ti]
		n := len(p.Cards)
		if n < 13 {
			continue
		}
		run := p.Cards[n-13:]
		if run[0].Rank() != King || !isRun(run, descSameSuit) {
			continue
		}
		for _, fi := range g.pilesOf(Foundation) {
			if len(g.Piles[fi].Cards) == 0 {
				g.Piles[fi].Cards = append(g.Piles[fi].Cards, run...)
				p.Cards = p.Cards[:n-13]
				return true
			}
		}
	}
	return false
}

// TapStock deals one card face-up onto every tableau column; refuses while
// any column is empty (the classic rule).
func (s *Spider) TapStock(g *Game) error {
	stock := &g.Piles[0]
	if len(stock.Cards) == 0 {
		return ErrStockEmpty
	}
	tableaus := g.pilesOf(Tableau)
	for _, ti := range tableaus {
		if len(g.Piles[ti].Cards) == 0 {
			return ErrEmptyColumn
		}
	}
	for _, ti := range tableaus {
		c := stock.Cards[len(stock.Cards)-1]
		stock.Cards = stock.Cards[:len(stock.Cards)-1]
		g.Piles[ti].Cards = append(g.Piles[ti].Cards, c.WithFaceUp(true))
	}
	return nil
}

func (s *Spider) Won(g *Game) bool {
	for _, fi := range g.pilesOf(Foundation) {
		if len(g.Piles[fi].Cards) != 13 {
			return false
		}
	}
	return true
}

func (s *Spider) SafeMoves(g *Game) []Move     { return nil } // runs remove themselves
func (s *Spider) AutoCompleteReady(*Game) bool { return false }
