package solitaire

// FreeCell: 8 tableau columns (7,7,7,7,6,6,6,6), 4 free cells, 4 foundations,
// no stock. Supermoves are engine-enforced and applied atomically.
type FreeCell struct{}

// NewFreeCell returns the FreeCell rules.
func NewFreeCell() *FreeCell { return &FreeCell{} }

func (f *FreeCell) ID() VariantID { return FreeCellV }
func (f *FreeCell) Name() string  { return "FreeCell" }

func (f *FreeCell) DeckSpec() DeckSpec {
	return DeckSpec{Decks: 1, Suits: []Suit{Spades, Hearts, Diamonds, Clubs}}
}

// Layout: 0-3=cells, 4-7=foundations, 8-15=tableau.
func (f *FreeCell) Layout() []PileKind {
	return []PileKind{
		Cell, Cell, Cell, Cell,
		Foundation, Foundation, Foundation, Foundation,
		Tableau, Tableau, Tableau, Tableau, Tableau, Tableau, Tableau, Tableau,
	}
}

func (f *FreeCell) Deal(g *Game, cards []Card) {
	for i, c := range cards {
		pile := &g.Piles[8+i%8]
		pile.Cards = append(pile.Cards, c.WithFaceUp(true))
	}
}

func (f *FreeCell) CanPickUp(g *Game, src, idx int) bool {
	p := &g.Piles[src]
	switch p.Kind {
	case Cell, Foundation:
		return idx == len(p.Cards)-1
	case Tableau:
		return isRun(p.Cards[idx:], descAltColor)
	default:
		return false
	}
}

func (f *FreeCell) CanDrop(g *Game, src, idx, dst int) bool {
	moving := g.Piles[src].Cards[idx:]
	d := &g.Piles[dst]
	switch d.Kind {
	case Cell:
		return len(moving) == 1 && len(d.Cards) == 0
	case Foundation:
		return len(moving) == 1 && foundationDrop(d, moving[0])
	case Tableau:
		if len(moving) > f.capacity(g, dst) {
			return false
		}
		top, ok := d.Top()
		if !ok {
			return true
		}
		return descAltColor(top, moving[0])
	default:
		return false
	}
}

// capacity is the supermove limit: (free cells + 1) doubled per empty
// tableau column — with the destination excluded when it is itself empty
// (the classic off-by-one).
func (f *FreeCell) capacity(g *Game, dst int) int {
	free := 0
	for _, ci := range g.pilesOf(Cell) {
		if len(g.Piles[ci].Cards) == 0 {
			free++
		}
	}
	empty := 0
	for _, ti := range g.pilesOf(Tableau) {
		if len(g.Piles[ti].Cards) == 0 && ti != dst {
			empty++
		}
	}
	return (free + 1) << empty
}

func (f *FreeCell) AfterMove(g *Game) {} // everything is already face-up

func (f *FreeCell) TapStock(g *Game) error { return ErrNoStock }

func (f *FreeCell) Won(g *Game) bool {
	for _, fi := range g.pilesOf(Foundation) {
		if len(g.Piles[fi].Cards) != 13 {
			return false
		}
	}
	return true
}

func (f *FreeCell) SafeMoves(g *Game) []Move { return safeFoundationSends(g) }

// AutoCompleteReady: every tableau column is a fully descending alternating
// run — trivially winnable from here.
func (f *FreeCell) AutoCompleteReady(g *Game) bool {
	if g.Won() {
		return false
	}
	for _, ti := range g.pilesOf(Tableau) {
		cards := g.Piles[ti].Cards
		if len(cards) > 0 && !isRun(cards, descAltColor) {
			return false
		}
	}
	return true
}
