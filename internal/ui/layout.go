package ui

import (
	"image"

	"github.com/richardwooding/patience/internal/solitaire"
)

// Canvas and table geometry.
const (
	W        = 960
	H        = 600
	toolbarH = 32
	tableTop = toolbarH + 8

	fanUp   = 18 // face-up vertical fan
	fanDown = 8  // face-down vertical fan
	fanMinU = 10 // compression floors
	fanMinD = 4
	wasteX  = 15 // waste horizontal fan
)

// pileLayout is one pile's screen geometry.
type pileLayout struct {
	pos     image.Point
	fanned  bool // vertical tableau fan
	isWaste bool // horizontal 3-card fan
}

// tableLayout maps every pile index to geometry for the current variant.
type tableLayout struct {
	piles []pileLayout
}

// layoutFor builds the pile geometry table for a variant's layout order.
func layoutFor(rules solitaire.Rules) tableLayout {
	kinds := rules.Layout()
	tl := tableLayout{piles: make([]pileLayout, len(kinds))}

	switch rules.ID() {
	case solitaire.FreeCellV:
		pitch := 87
		left := (W - (8*CardW + 7*(pitch-CardW))) / 2
		for i := range 4 { // cells
			tl.piles[i] = pileLayout{pos: image.Pt(left+i*pitch, tableTop)}
		}
		for i := range 4 { // foundations
			tl.piles[4+i] = pileLayout{pos: image.Pt(left+(4+i)*pitch, tableTop)}
		}
		for i := range 8 { // tableau
			tl.piles[8+i] = pileLayout{pos: image.Pt(left+i*pitch, tableTop+CardH+12), fanned: true}
		}
	case solitaire.Spider1, solitaire.Spider2, solitaire.Spider4:
		pitch := 79
		left := (W - (10*CardW + 9*(pitch-CardW))) / 2
		tl.piles[0] = pileLayout{pos: image.Pt(W-CardW-16, tableTop)} // stock
		for i := range 8 {                                            // foundations, compact row
			tl.piles[1+i] = pileLayout{pos: image.Pt(16+i*40, tableTop)}
		}
		for i := range 10 { // tableau
			tl.piles[9+i] = pileLayout{pos: image.Pt(left+i*pitch, tableTop+CardH+12), fanned: true}
		}
	default: // Klondike
		pitch := 96
		left := (W - (7*CardW + 6*(pitch-CardW))) / 2
		tl.piles[0] = pileLayout{pos: image.Pt(left, tableTop)}                      // stock
		tl.piles[1] = pileLayout{pos: image.Pt(left+pitch, tableTop), isWaste: true} // waste
		for i := range 4 {                                                           // foundations
			tl.piles[2+i] = pileLayout{pos: image.Pt(left+(3+i)*pitch, tableTop)}
		}
		for i := range 7 { // tableau
			tl.piles[6+i] = pileLayout{pos: image.Pt(left+i*pitch, tableTop+CardH+12), fanned: true}
		}
	}
	return tl
}

// positions computes every card's screen position in pile pi.
func (tl tableLayout) positions(g *solitaire.Game, pi int) []image.Point {
	p := &g.Piles[pi]
	li := tl.piles[pi]
	switch {
	case li.fanned:
		return fannedPositions(p, li)
	case li.isWaste:
		return wastePositions(p, li)
	default:
		out := make([]image.Point, len(p.Cards))
		for i := range p.Cards {
			out[i] = li.pos
		}
		return out
	}
}

// fannedPositions lays a tableau column vertically, compressing the fan so
// tall Spider columns always fit above the bottom edge.
func fannedPositions(p *solitaire.Pile, li pileLayout) []image.Point {
	upN, downN := 0, 0
	for _, c := range p.Cards {
		if c.FaceUp() {
			upN++
		} else {
			downN++
		}
	}
	fu, fd := fanUp, fanDown
	avail := H - li.pos.Y - CardH - 8
	if need := downN*fd + upN*fu; need > avail && need > 0 {
		scale := float64(avail) / float64(need)
		fu = max(int(float64(fu)*scale), fanMinU)
		fd = max(int(float64(fd)*scale), fanMinD)
	}
	out := make([]image.Point, len(p.Cards))
	y := li.pos.Y
	for i, c := range p.Cards {
		out[i] = image.Pt(li.pos.X, y)
		if c.FaceUp() {
			y += fu
		} else {
			y += fd
		}
	}
	return out
}

// wastePositions fans the last up-to-3 waste cards horizontally.
func wastePositions(p *solitaire.Pile, li pileLayout) []image.Point {
	out := make([]image.Point, len(p.Cards))
	start := max(len(p.Cards)-3, 0)
	for i := range p.Cards {
		x := li.pos.X
		if i >= start {
			x += (i - start) * wasteX
		}
		out[i] = image.Pt(x, li.pos.Y)
	}
	return out
}

// cardRect is card i's full hit rectangle.
func (tl tableLayout) cardRect(g *solitaire.Game, pi, i int) image.Rectangle {
	pos := tl.positions(g, pi)[i]
	return image.Rect(pos.X, pos.Y, pos.X+CardW, pos.Y+CardH)
}

// slotRect is the pile's empty-slot rectangle.
func (tl tableLayout) slotRect(pi int) image.Rectangle {
	pos := tl.piles[pi].pos
	return image.Rect(pos.X, pos.Y, pos.X+CardW, pos.Y+CardH)
}

// hitCard finds the topmost card under pt (inflated for touch), returning
// pile and card index. ok=false when nothing is hit.
func (tl tableLayout) hitCard(g *solitaire.Game, pt image.Point, inflate int) (pile, idx int, ok bool) {
	best := -1
	for pi := range g.Piles {
		for i := range g.Piles[pi].Cards {
			r := tl.cardRect(g, pi, i).Inset(-inflate)
			if pt.In(r) && i >= best {
				// later piles/cards win only via higher card index within
				// a pile; across piles rects don't meaningfully overlap
				pile, idx, ok = pi, i, true
				best = i
			}
		}
	}
	return pile, idx, ok
}

// hitPile finds the pile whose slot or cards are under pt (for empty-pile
// drops and stock taps).
func (tl tableLayout) hitPile(g *solitaire.Game, pt image.Point) (int, bool) {
	for pi := range g.Piles {
		if len(g.Piles[pi].Cards) == 0 && pt.In(tl.slotRect(pi)) {
			return pi, true
		}
	}
	if pi, _, ok := tl.hitCard(g, pt, 0); ok {
		return pi, true
	}
	return 0, false
}

// dropTarget picks the destination pile whose area overlaps the dragged
// card rect the most (better than pointer-based targeting on touch),
// filtered by rules legality.
func (tl tableLayout) dropTarget(g *solitaire.Game, drag image.Rectangle, src, idx int) (int, bool) {
	bestPile, bestArea := -1, 0
	for pi := range g.Piles {
		if pi == src {
			continue
		}
		var zone image.Rectangle
		if n := len(g.Piles[pi].Cards); n > 0 {
			zone = tl.cardRect(g, pi, n-1)
		} else {
			zone = tl.slotRect(pi)
		}
		sect := zone.Intersect(drag)
		area := sect.Dx() * sect.Dy()
		if area <= 0 {
			continue
		}
		if !g.Rules.CanDrop(g, src, idx, pi) {
			continue
		}
		if area > bestArea {
			bestPile, bestArea = pi, area
		}
	}
	return bestPile, bestPile >= 0
}
