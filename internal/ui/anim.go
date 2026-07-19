package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Flight animates one card sprite between two points; the model is already
// updated, so flights are pure presentation. While a flight targets a slot,
// that slot's top cards are hidden via the hide map in the scene.
type flight struct {
	sprite   *ebiten.Image
	from, to image.Point
	t, dur   float64
	pile     int // destination pile whose top card is hidden until arrival
	depth    int // how many cards this flight covers (stacked sends)
}

// step advances the flight; done when t >= dur.
func (f *flight) step() bool {
	f.t++
	return f.t >= f.dur
}

// pos is the eased current position (ease-out quadratic).
func (f *flight) pos() image.Point {
	p := f.t / f.dur
	if p > 1 {
		p = 1
	}
	e := 1 - (1-p)*(1-p)
	return image.Pt(
		f.from.X+int(float64(f.to.X-f.from.X)*e),
		f.from.Y+int(float64(f.to.Y-f.from.Y)*e),
	)
}
