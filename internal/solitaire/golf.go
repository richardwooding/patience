package solitaire

// Golf: 7 tableau columns of 5 face-up cards, one foundation, and a stock
// dealt one card at a time onto the foundation. A tableau top may go to the
// foundation when it is one rank above or below the current top (no wrap —
// King and Ace are not adjacent). Clear every column to win.
type Golf struct{}

// NewGolf returns the Golf rules.
func NewGolf() *Golf { return &Golf{} }

func (g *Golf) ID() VariantID { return GolfV }
func (g *Golf) Name() string  { return "Golf" }

func (g *Golf) DeckSpec() DeckSpec {
	return DeckSpec{Decks: 1, Suits: []Suit{Spades, Hearts, Diamonds, Clubs}}
}

// Layout: 0=stock, 1=foundation, 2-8=tableau.
func (g *Golf) Layout() []PileKind {
	return []PileKind{
		Stock, Foundation,
		Tableau, Tableau, Tableau, Tableau, Tableau, Tableau, Tableau,
	}
}

func (g *Golf) Deal(game *Game, cards []Card) {
	next := 0
	for col := range 7 {
		p := &game.Piles[2+col]
		for range 5 {
			p.Cards = append(p.Cards, cards[next].WithFaceUp(true))
			next++
		}
	}
	// One card seeds the foundation face-up; the rest are the stock.
	game.Piles[1].Cards = append(game.Piles[1].Cards, cards[next].WithFaceUp(true))
	next++
	game.Piles[0].Cards = append(game.Piles[0].Cards, cards[next:]...)
}

func (g *Golf) CanPickUp(game *Game, src, idx int) bool {
	p := &game.Piles[src]
	return p.Kind == Tableau && idx == len(p.Cards)-1
}

func (g *Golf) CanDrop(game *Game, src, idx, dst int) bool {
	if game.Piles[dst].Kind != Foundation {
		return false
	}
	top, ok := game.Piles[dst].Top()
	if !ok {
		return false // the foundation is seeded at deal and never empties
	}
	return rankDiff(game.Piles[src].Cards[idx].Rank(), top.Rank()) == 1
}

func (g *Golf) AfterMove(*Game) {} // nothing cascades

// TapStock deals one card face-up onto the foundation. No recycle.
func (g *Golf) TapStock(game *Game) error {
	stock := &game.Piles[0]
	if len(stock.Cards) == 0 {
		return ErrStockEmpty
	}
	c := stock.Cards[len(stock.Cards)-1]
	stock.Cards = stock.Cards[:len(stock.Cards)-1]
	game.Piles[1].Cards = append(game.Piles[1].Cards, c.WithFaceUp(true))
	return nil
}

func (g *Golf) Won(game *Game) bool {
	for _, ti := range game.pilesOf(Tableau) {
		if len(game.Piles[ti].Cards) > 0 {
			return false
		}
	}
	return true
}

func (g *Golf) SafeMoves(*Game) []Move       { return nil }
func (g *Golf) AutoCompleteReady(*Game) bool { return false }

// rankDiff is the absolute difference between two ranks.
func rankDiff(a, b Rank) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}
