package ui

import (
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var face = text.NewGoXFace(bitmapfont.Face)

const (
	glyphW = 6
	glyphH = 13
)

func drawText(dst *ebiten.Image, s string, x, y float64, clr interface{ RGBA() (r, g, b, a uint32) }, scale float64) {
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	r, g, b, a := clr.RGBA()
	op.ColorScale.Scale(float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff)
	text.Draw(dst, s, face, op)
}

func textWidth(s string, scale float64) float64 {
	return float64(len(s)) * glyphW * scale
}
