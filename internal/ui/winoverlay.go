package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/richardwooding/patience/internal/solitaire"
	"github.com/richardwooding/patience/internal/stats"
)

// winOverlay celebrates and offers the next deal.
type winOverlay struct {
	g       *solitaire.Game
	ptr     pointer
	btnNew  button
	btnMenu button
}

func newWinOverlay(g *solitaire.Game) *winOverlay {
	return &winOverlay{
		g:       g,
		btnNew:  button{x: W/2 - 130, y: 330, w: 120, h: 36, label: "deal again"},
		btnMenu: button{x: W/2 + 10, y: 330, w: 120, h: 36, label: "menu"},
	}
}

func (w *winOverlay) Update(g *Game) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyN) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.scene = newTableScene(w.g.Rules, newSeed())
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.scene = newMenuScene()
		return nil
	}
	pos, pressed, _, _ := w.ptr.state()
	if pressed {
		pts := []image.Point{pos}
		switch {
		case w.btnNew.hit(pts):
			g.scene = newTableScene(w.g.Rules, newSeed())
		case w.btnMenu.hit(pts):
			g.scene = newMenuScene()
		}
	}
	return nil
}

func (w *winOverlay) Draw(dst *ebiten.Image) {
	title := "YOU WON"
	drawText(dst, title, (W-textWidth(title, 6))/2, 120, colAccent, 6)

	line := fmt.Sprintf("%s — %d moves", w.g.Rules.Name(), w.g.MoveCount)
	drawText(dst, line, (W-textWidth(line, 2))/2, 220, colText, 2)

	st := stats.Get(string(w.g.Rules.ID()))
	rec := fmt.Sprintf("won %d of %d · best %d moves", st.Won, st.Played, st.BestMoves)
	drawText(dst, rec, (W-textWidth(rec, 1))/2, 260, colDim, 1)

	w.btnNew.draw(dst, true)
	w.btnMenu.draw(dst, false)
}
