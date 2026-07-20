package solitaire

import (
	"errors"
	"fmt"
)

// PileKind is a pile's role; the Rules' Layout() fixes each variant's pile
// order, so indices are stable and the UI maps (Kind, ordinal) to geometry.
type PileKind uint8

const (
	Stock PileKind = iota
	Waste
	Foundation
	Tableau
	Cell
)

// Pile is an ordered stack of cards; index 0 is the bottom.
type Pile struct {
	Kind  PileKind
	Cards []Card
}

// Top returns the top card, or false when empty.
func (p *Pile) Top() (Card, bool) {
	if len(p.Cards) == 0 {
		return 0, false
	}
	return p.Cards[len(p.Cards)-1], true
}

// Move transfers Piles[Src].Cards[Idx:] onto Piles[Dst].
type Move struct {
	Src, Idx, Dst int
}

// Errors the UI turns into toasts.
var (
	ErrIllegalMove = errors.New("illegal move")
	ErrNoStock     = errors.New("this variant has no stock")
	ErrEmptyColumn = errors.New("fill every empty column before dealing")
	ErrStockEmpty  = errors.New("stock is empty")
)

// snapshot captures everything Apply/TapStock can change.
type snapshot struct {
	piles     []Pile
	moveCount int
	recycles  int
}

// Game is one deal in progress.
type Game struct {
	Rules     Rules
	Piles     []Pile
	Seed      uint64
	MoveCount int
	Recycles  int // Klondike stock pass-throughs

	undo []snapshot
}

// New deals a fresh game for the rules and seed.
func New(r Rules, seed uint64) *Game {
	g := &Game{Rules: r, Seed: seed}
	for _, kind := range r.Layout() {
		g.Piles = append(g.Piles, Pile{Kind: kind})
	}
	r.Deal(g, Shuffled(r.DeckSpec(), seed))
	return g
}

// pilesOf returns the indices of every pile of the given kind, in layout order.
func (g *Game) pilesOf(kind PileKind) []int {
	var out []int
	for i := range g.Piles {
		if g.Piles[i].Kind == kind {
			out = append(out, i)
		}
	}
	return out
}

// Legal reports whether m is currently allowed.
func (g *Game) Legal(m Move) bool {
	if m.Src < 0 || m.Src >= len(g.Piles) || m.Dst < 0 || m.Dst >= len(g.Piles) || m.Src == m.Dst {
		return false
	}
	if m.Idx < 0 || m.Idx >= len(g.Piles[m.Src].Cards) {
		return false
	}
	return g.Rules.CanPickUp(g, m.Src, m.Idx) && g.Rules.CanDrop(g, m.Src, m.Idx, m.Dst)
}

// Apply executes m (validated), records undo, and runs the variant's
// AfterMove hook (exposure flips, Spider run removal).
func (g *Game) Apply(m Move) error {
	if !g.Legal(m) {
		return fmt.Errorf("%w: %+v", ErrIllegalMove, m)
	}
	g.pushUndo()
	src, dst := &g.Piles[m.Src], &g.Piles[m.Dst]
	moved := src.Cards[m.Idx:]
	dst.Cards = append(dst.Cards, moved...)
	src.Cards = src.Cards[:m.Idx]
	g.MoveCount++
	g.Rules.AfterMove(g)
	return nil
}

// TapStock runs the variant's stock action (draw, recycle, deal a row).
func (g *Game) TapStock() error {
	g.pushUndo()
	if err := g.Rules.TapStock(g); err != nil {
		g.popUndoDiscard()
		return err
	}
	g.MoveCount++
	g.Rules.AfterMove(g)
	return nil
}

// Won reports whether the game is complete.
func (g *Game) Won() bool { return g.Rules.Won(g) }

// CanUndo reports whether an undo step exists.
func (g *Game) CanUndo() bool { return len(g.undo) > 0 }

// Undo restores the state before the most recent Apply/TapStock.
func (g *Game) Undo() bool {
	if len(g.undo) == 0 {
		return false
	}
	s := g.undo[len(g.undo)-1]
	g.undo = g.undo[:len(g.undo)-1]
	g.Piles = s.piles
	g.MoveCount = s.moveCount
	g.Recycles = s.recycles
	return true
}

func (g *Game) pushUndo() {
	s := snapshot{
		piles:     make([]Pile, len(g.Piles)),
		moveCount: g.MoveCount,
		recycles:  g.Recycles,
	}
	for i, p := range g.Piles {
		s.piles[i] = Pile{Kind: p.Kind, Cards: append([]Card(nil), p.Cards...)}
	}
	g.undo = append(g.undo, s)
}

func (g *Game) popUndoDiscard() {
	g.undo = g.undo[:len(g.undo)-1]
}

// flipExposed turns up the top card of every tableau pile — the shared
// AfterMove behavior for Klondike and Spider.
func (g *Game) flipExposed() {
	for i := range g.Piles {
		p := &g.Piles[i]
		if p.Kind != Tableau {
			continue
		}
		if top, ok := p.Top(); ok && !top.FaceUp() {
			p.Cards[len(p.Cards)-1] = top.WithFaceUp(true)
		}
	}
}

// Hint returns a suggested useful move, preferring (in order) moves that
// expose a face-down card, safe foundation sends, and builds onto a non-empty
// pile, falling back to any legal move. It reports false when no card move
// exists (the stock, if any, may still be tappable — see AnyLegalMove).
func (g *Game) Hint() (Move, bool) {
	moves := g.legalMoves()
	for _, m := range moves {
		if g.exposesFaceDown(m) {
			return m, true
		}
	}
	if s, ok := g.firstSafeSend(); ok {
		return s, true
	}
	for _, m := range moves {
		if g.buildsOnPile(m) {
			return m, true
		}
	}
	if len(moves) > 0 {
		return moves[0], true
	}
	return Move{}, false
}

// legalMoves lists every legal card move on the board (no stock taps).
func (g *Game) legalMoves() []Move {
	var moves []Move
	for src := range g.Piles {
		for idx := range g.Piles[src].Cards {
			if g.Rules.CanPickUp(g, src, idx) {
				moves = append(moves, g.dropsFor(src, idx)...)
			}
		}
	}
	return moves
}

// dropsFor lists the legal destinations for the run at Piles[src].Cards[idx:].
func (g *Game) dropsFor(src, idx int) []Move {
	var moves []Move
	for dst := range g.Piles {
		if dst != src && g.Rules.CanDrop(g, src, idx, dst) {
			moves = append(moves, Move{Src: src, Idx: idx, Dst: dst})
		}
	}
	return moves
}

// exposesFaceDown reports whether m uncovers a face-down tableau card.
func (g *Game) exposesFaceDown(m Move) bool {
	src := &g.Piles[m.Src]
	return src.Kind == Tableau && m.Idx > 0 && !src.Cards[m.Idx-1].FaceUp()
}

// firstSafeSend returns the first safe foundation send, computing the
// board-wide safe set once (it does not depend on any single candidate move).
func (g *Game) firstSafeSend() (Move, bool) {
	// safeFoundationSends uses the generic suit-ascending foundation rule, which
	// does not hold for every variant (Pyramid's foundation takes only Kings,
	// Golf's builds by rank), so confirm each candidate is actually legal before
	// hinting it.
	for _, s := range safeFoundationSends(g) {
		if g.Legal(s) {
			return s, true
		}
	}
	return Move{}, false
}

// buildsOnPile reports whether m stacks onto a non-empty, non-foundation pile
// (consolidation), as opposed to moving into empty space.
func (g *Game) buildsOnPile(m Move) bool {
	dst := &g.Piles[m.Dst]
	return dst.Kind != Foundation && len(dst.Cards) > 0
}

// AnyLegalMove reports whether any pickup/drop or stock action remains.
func (g *Game) AnyLegalMove() bool {
	for src := range g.Piles {
		for idx := range g.Piles[src].Cards {
			if !g.Rules.CanPickUp(g, src, idx) {
				continue
			}
			for dst := range g.Piles {
				if dst != src && g.Rules.CanDrop(g, src, idx, dst) {
					return true
				}
			}
		}
	}
	// A tappable stock also counts as a move.
	probe := g.clone()
	return probe.Rules.TapStock(probe) == nil
}

// clone deep-copies the game (used for non-destructive probes).
func (g *Game) clone() *Game {
	c := &Game{Rules: g.Rules, Seed: g.Seed, MoveCount: g.MoveCount, Recycles: g.Recycles}
	c.Piles = make([]Pile, len(g.Piles))
	for i, p := range g.Piles {
		c.Piles[i] = Pile{Kind: p.Kind, Cards: append([]Card(nil), p.Cards...)}
	}
	return c
}
