package ui

import (
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/patience/internal/solitaire"
	"github.com/richardwooding/patience/internal/stats"
)

const (
	menuListY = 200
	menuRowH  = 52
)

// newSeed picks a fresh deal seed.
func newSeed() uint64 { return rand.Uint64() }

// menuScene picks a variant.
type menuScene struct {
	selected int
	ptr      pointer
}

func newMenuScene() *menuScene {
	ensureSprites()
	return &menuScene{}
}

func (m *menuScene) Update(g *Game) error {
	variants := solitaire.Variants()
	if m.handleKeys(g, variants) {
		return nil
	}
	pos, pressed, _, _ := m.ptr.state()
	if !pressed {
		return nil
	}
	i, ok := menuRowAt(pos, len(variants))
	if !ok {
		return nil
	}
	if m.selected == i {
		g.scene = newTableScene(variants[i], newSeed())
		return nil
	}
	m.selected = i
	return nil
}

// handleKeys moves the selection and deals on enter/space; reports whether
// the scene changed.
func (m *menuScene) handleKeys(g *Game, variants []solitaire.Rules) bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && m.selected < len(variants)-1 {
		m.selected++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && m.selected > 0 {
		m.selected--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.scene = newTableScene(variants[m.selected], newSeed())
		return true
	}
	return false
}

// menuRowAt maps a point to a variant-list row.
func menuRowAt(pos image.Point, n int) (int, bool) {
	if pos.X < 200 || pos.X >= W-200 || pos.Y < menuListY {
		return 0, false
	}
	i := (pos.Y - menuListY) / menuRowH
	if i >= n {
		return 0, false
	}
	return i, true
}

func (m *menuScene) Draw(dst *ebiten.Image) {
	title := "PATIENCE"
	drawText(dst, title, (W-textWidth(title, 6))/2, 48, colAccent, 6)
	sub := "retro solitaire — 100% Go, in your browser"
	drawText(dst, sub, (W-textWidth(sub, 2))/2, 136, colDim, 2)

	for i, v := range solitaire.Variants() {
		y := float64(menuListY + i*menuRowH)
		if i == m.selected {
			vector.FillRect(dst, 200, float32(y)-6, W-400, menuRowH-8, colPanel, false)
			vector.StrokeRect(dst, 200, float32(y)-6, W-400, menuRowH-8, 1, colAccent, false)
			drawText(dst, "▶", 220, y+6, colAccent, 2)
		}
		drawText(dst, v.Name(), 252, y+6, colText, 2)
		st := stats.Get(string(v.ID()))
		if st.Played > 0 {
			line := fmt.Sprintf("won %d/%d", st.Won, st.Played)
			if st.BestMoves > 0 {
				line += fmt.Sprintf(" · best %d moves", st.BestMoves)
			}
			drawText(dst, line, W-220-textWidth(line, 1), y+12, colDim, 1)
		}
	}

	foot := "tap/enter to deal · in game: u undo · r restart · a auto-finish · esc menu"
	drawText(dst, foot, (W-textWidth(foot, 1))/2, H-32, colDimmer, 1)
}
