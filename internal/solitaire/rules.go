package solitaire

// VariantID names a playable configuration.
type VariantID string

const (
	Klondike1 VariantID = "klondike-1"
	Klondike3 VariantID = "klondike-3"
	FreeCellV VariantID = "freecell"
	Spider1   VariantID = "spider-1"
	Spider2   VariantID = "spider-2"
	Spider4   VariantID = "spider-4"
)

// Rules is the variant abstraction: pure predicates plus a few behavioral
// hooks. The generic Game does all mutation; implementations stay small.
type Rules interface {
	ID() VariantID
	Name() string
	DeckSpec() DeckSpec
	Layout() []PileKind
	Deal(g *Game, cards []Card)
	CanPickUp(g *Game, src, idx int) bool
	CanDrop(g *Game, src, idx, dst int) bool
	AfterMove(g *Game)
	TapStock(g *Game) error
	Won(g *Game) bool
	SafeMoves(g *Game) []Move
	AutoCompleteReady(g *Game) bool
}

// Variants lists every playable configuration in menu order.
func Variants() []Rules {
	return []Rules{
		NewKlondike(1), NewKlondike(3),
		NewFreeCell(),
		NewSpider(1), NewSpider(2), NewSpider(4),
	}
}

// --- shared predicates ---

// descAltColor reports whether under can sit on over in an alternating-color
// descending run (over = K♠, under = Q♥ → true).
func descAltColor(over, under Card) bool {
	return over.Rank() == under.Rank()+1 && over.IsRed() != under.IsRed()
}

// descSameSuit reports a same-suit descending adjacency.
func descSameSuit(over, under Card) bool {
	return over.Rank() == under.Rank()+1 && over.Suit() == under.Suit()
}

// isRun reports whether cards form a face-up run under the adjacency rule.
func isRun(cards []Card, adjacent func(over, under Card) bool) bool {
	for i, c := range cards {
		if !c.FaceUp() {
			return false
		}
		if i > 0 && !adjacent(cards[i-1], c) {
			return false
		}
	}
	return len(cards) > 0
}

// foundationDrop reports whether a single card c may go on foundation pile p
// (same suit ascending from the ace).
func foundationDrop(p *Pile, c Card) bool {
	top, ok := p.Top()
	if !ok {
		return c.Rank() == Ace
	}
	return c.Suit() == top.Suit() && c.Rank() == top.Rank()+1
}

// safeFoundationSends returns foundation moves that can never hurt: the card
// is rank ≤ 2, or both opposite-color foundations have reached rank-1.
func safeFoundationSends(g *Game) []Move {
	var moves []Move
	foundations := g.pilesOf(Foundation)
	// A color with any empty/low foundation floors at 0, keeping sends safe.
	redFloor, blackFloor := foundationFloor(g, true), foundationFloor(g, false)

	for src := range g.Piles {
		p := &g.Piles[src]
		if p.Kind != Tableau && p.Kind != Waste && p.Kind != Cell {
			continue
		}
		top, ok := p.Top()
		if !ok || !top.FaceUp() {
			continue
		}
		opposite := blackFloor
		if !top.IsRed() {
			opposite = redFloor
		}
		if top.Rank() > 2 && top.Rank() > opposite+1 {
			continue // sending it could strand a card that needs it
		}
		for _, fi := range foundations {
			if foundationDrop(&g.Piles[fi], top) {
				moves = append(moves, Move{Src: src, Idx: len(p.Cards) - 1, Dst: fi})
				break
			}
		}
	}
	return moves
}

// foundationFloor is the lowest foundation rank among the given color
// (0 when any foundation of that color is still empty). Two foundations per
// color in a single-deck game.
func foundationFloor(g *Game, red bool) Rank {
	count := 0
	floor := King
	for _, fi := range g.pilesOf(Foundation) {
		top, ok := g.Piles[fi].Top()
		if !ok {
			continue
		}
		if top.IsRed() == red {
			count++
			if top.Rank() < floor {
				floor = top.Rank()
			}
		}
	}
	if count < 2 {
		return 0 // an empty same-color foundation means floor is zero
	}
	return floor
}
