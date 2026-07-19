package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// button is the house minimal tap target (chipdeck pattern).
type button struct {
	x, y, w, h float32
	label      string
}

func (b button) hit(pts []image.Point) bool {
	for _, p := range pts {
		if float32(p.X) >= b.x && float32(p.X) < b.x+b.w &&
			float32(p.Y) >= b.y && float32(p.Y) < b.y+b.h {
			return true
		}
	}
	return false
}

func (b button) draw(dst *ebiten.Image, accent bool) {
	vector.FillRect(dst, b.x, b.y, b.w, b.h, colPanel, false)
	edge, txt := colPanelEdge, colText
	if accent {
		edge, txt = colAccent, colAccent
	}
	vector.StrokeRect(dst, b.x, b.y, b.w, b.h, 1, edge, false)
	tw := textWidth(b.label, 1)
	drawText(dst, b.label, float64(b.x)+(float64(b.w)-tw)/2, float64(b.y)+(float64(b.h)-glyphH)/2, txt, 1)
}

// accentOutline draws a 2px accent outline around r (the drop highlight).
func accentOutline(dst *ebiten.Image, r image.Rectangle) {
	vector.StrokeRect(dst, float32(r.Min.X)-1, float32(r.Min.Y)-1, float32(r.Dx())+2, float32(r.Dy())+2, 2, colAccent, false)
}
