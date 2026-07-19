package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/patience/internal/solitaire"
)

// Card sprites at the classic Windows 3.x CARDS.DLL size.
const (
	CardW = 71
	CardH = 96
)

var (
	faceSprites [4][14]*ebiten.Image // [suit][rank]
	backSprite  *ebiten.Image
	slotSprite  *ebiten.Image
)

// ensureSprites builds all card art programmatically, once.
func ensureSprites() {
	if backSprite != nil {
		return
	}
	for s := solitaire.Spades; s <= solitaire.Clubs; s++ {
		for r := solitaire.Ace; r <= solitaire.King; r++ {
			faceSprites[s][r] = buildFace(solitaire.NewCard(s, r))
		}
	}
	backSprite = buildBack()
	slotSprite = buildSlot()
}

// sprite returns the image for a card (face or back).
func sprite(c solitaire.Card) *ebiten.Image {
	if !c.FaceUp() {
		return backSprite
	}
	return faceSprites[c.Suit()][c.Rank()]
}

// cardBase fills a rounded card body with a border.
func cardBase(img *ebiten.Image, fill, border color.RGBA) {
	img.Fill(fill)
	// chunky 2px retro corner knockout
	for _, x := range []int{0, CardW - 2} {
		for _, y := range []int{0, CardH - 2} {
			vector.FillRect(img, float32(x), float32(y), 2, 2, colBG, false)
		}
	}
	vector.StrokeRect(img, 0.5, 0.5, CardW-1, CardH-1, 1, border, false)
}

func buildFace(c solitaire.Card) *ebiten.Image {
	img := ebiten.NewImage(CardW, CardH)
	cardBase(img, colCardFace, colCardInk)

	ink := colCardInk
	if c.IsRed() {
		ink = colCardRed
	}

	label := c.RankName()
	glyph := string(c.SuitRune())

	// corner indices, top-left and bottom-right (rotated by symmetry)
	drawText(img, label, 5, 3, ink, 1)
	drawText(img, glyph, 5, 15, ink, 1)
	drawText(img, label, float64(CardW)-5-textWidth(label, 1), float64(CardH)-31, ink, 1)
	drawText(img, glyph, float64(CardW)-5-textWidth(glyph, 1), float64(CardH)-19, ink, 1)

	if c.Rank() >= solitaire.Jack {
		// court cards: framed big rank letter
		vector.StrokeRect(img, 16, 24, CardW-32, CardH-48, 1, ink, false)
		vector.StrokeRect(img, 19, 27, CardW-38, CardH-54, 1, ink, false)
		drawText(img, label, (CardW-textWidth(label, 2))/2, float64(CardH)/2-glyphH, ink, 2)
	} else {
		// one big center pip
		blitPip(img, int(c.Suit()), (CardW-32)/2, (CardH-32)/2, ink)
	}
	return img
}

// blitPip draws a 16x16 pip mask at 2x scale.
func blitPip(img *ebiten.Image, suit, x, y int, ink color.RGBA) {
	mask := pipMasks[suit]
	for row := range 16 {
		bits := mask[row]
		for col := range 16 {
			if bits&(1<<(15-col)) != 0 {
				vector.FillRect(img, float32(x+col*2), float32(y+row*2), 2, 2, ink, false)
			}
		}
	}
}

func buildBack() *ebiten.Image {
	img := ebiten.NewImage(CardW, CardH)
	cardBase(img, colCardBack, colAccentDim)
	// purple diamond lattice on an 8px grid
	for y := 4; y < CardH-4; y += 8 {
		for x := 4; x < CardW-4; x += 8 {
			vector.StrokeLine(img, float32(x), float32(y+4), float32(x+4), float32(y), 1, colAccentDim, false)
			vector.StrokeLine(img, float32(x+4), float32(y), float32(x+8), float32(y+4), 1, colAccentDim, false)
			vector.StrokeLine(img, float32(x+8), float32(y+4), float32(x+4), float32(y+8), 1, colAccentDim, false)
			vector.StrokeLine(img, float32(x+4), float32(y+8), float32(x), float32(y+4), 1, colAccentDim, false)
		}
	}
	vector.StrokeRect(img, 3.5, 3.5, CardW-7, CardH-7, 1, colAccent, false)
	return img
}

func buildSlot() *ebiten.Image {
	img := ebiten.NewImage(CardW, CardH)
	vector.StrokeRect(img, 0.5, 0.5, CardW-1, CardH-1, 1, colDimmer, false)
	return img
}
