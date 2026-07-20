package ui

import (
	"fmt"
	"image"
	"math/rand/v2"
	"time"

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

// menuScene picks a variant, casual or daily.
type menuScene struct {
	selected int
	daily    bool
	today    int
	btnDaily button
	ptr      pointer
}

func newMenuScene() *menuScene {
	ensureSprites()
	return &menuScene{
		today:    solitaire.DayNumber(time.Now()),
		btnDaily: button{x: W/2 - 90, y: 164, w: 180, h: 28, label: "daily deal: off"},
	}
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
	if m.btnDaily.hit([]image.Point{pos}) {
		m.daily = !m.daily
		return nil
	}
	i, ok := menuRowAt(pos, len(variants))
	if !ok {
		return nil
	}
	if m.selected == i {
		m.start(g, variants[i])
		return nil
	}
	m.selected = i
	return nil
}

// start opens the selected variant, daily or casual.
func (m *menuScene) start(g *Game, v solitaire.Rules) {
	if m.daily {
		g.scene = newDailyScene(v, m.today)
		return
	}
	g.scene = newTableScene(v, newSeed())
}

// handleKeys moves the selection, toggles daily, and deals on enter/space;
// reports whether the scene changed.
func (m *menuScene) handleKeys(g *Game, variants []solitaire.Rules) bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && m.selected < len(variants)-1 {
		m.selected++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && m.selected > 0 {
		m.selected--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		m.daily = !m.daily
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.start(g, variants[m.selected])
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

	m.drawDailyToggle(dst)

	for i, v := range solitaire.Variants() {
		y := float64(menuListY + i*menuRowH)
		if i == m.selected {
			vector.FillRect(dst, 200, float32(y)-6, W-400, menuRowH-8, colPanel, false)
			vector.StrokeRect(dst, 200, float32(y)-6, W-400, menuRowH-8, 1, colAccent, false)
			drawText(dst, "▶", 220, y+6, colAccent, 2)
		}
		drawText(dst, v.Name(), 252, y+6, colText, 2)
		if line := m.rowStat(string(v.ID())); line != "" {
			drawText(dst, line, W-220-textWidth(line, 1), y+12, colDim, 1)
		}
	}

	foot := "tap/enter to deal · d daily · in game: u undo · h hint · a auto-finish · esc menu"
	drawText(dst, foot, (W-textWidth(foot, 1))/2, H-32, colDimmer, 1)
}

func (m *menuScene) drawDailyToggle(dst *ebiten.Image) {
	m.btnDaily.label = "daily deal: off"
	if m.daily {
		m.btnDaily.label = fmt.Sprintf("daily deal: on · #%d", m.today)
	}
	m.btnDaily.draw(dst, m.daily)
}

// rowStat is the right-aligned per-variant line: daily streak in daily mode,
// else the casual win record.
func (m *menuScene) rowStat(id string) string {
	if m.daily {
		d := stats.GetDaily(id)
		if d.SolvedToday(m.today) {
			return fmt.Sprintf("solved today · streak %d", d.Streak)
		}
		if d.Wins > 0 {
			return fmt.Sprintf("streak %d · best streak %d", d.Streak, d.MaxStreak)
		}
		return "play today's deal"
	}
	st := stats.Get(id)
	if st.Played == 0 {
		return ""
	}
	if st.BestMoves > 0 {
		return fmt.Sprintf("won %d/%d · best %d moves", st.Won, st.Played, st.BestMoves)
	}
	return fmt.Sprintf("won %d/%d", st.Won, st.Played)
}
