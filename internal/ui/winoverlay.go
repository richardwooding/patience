package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/richardwooding/patience/internal/share"
	"github.com/richardwooding/patience/internal/solitaire"
	"github.com/richardwooding/patience/internal/stats"
)

// winInfo carries a finished game's result to the overlay.
type winInfo struct {
	g         *solitaire.Game
	daily     bool
	day       int
	undos     int
	hints     int
	streak    int
	maxStreak int
	shareURL  string
}

// winOverlay celebrates and offers the next step. Casual games get "deal
// again"; daily games get a "share" button that copies the result blurb.
type winOverlay struct {
	info      winInfo
	ptr       pointer
	btnMain   button // "deal again" (casual only)
	btnShare  button // "share" (daily only)
	btnMenu   button
	copiedTTL int
}

func newWinOverlay(info winInfo) *winOverlay {
	w := &winOverlay{
		info:    info,
		btnMenu: button{x: W/2 + 10, y: 330, w: 120, h: 36, label: "menu"},
	}
	if info.daily {
		w.btnShare = button{x: W/2 - 130, y: 330, w: 120, h: 36, label: "share"}
	} else {
		w.btnMain = button{x: W/2 - 130, y: 330, w: 120, h: 36, label: "deal again"}
	}
	return w
}

func (w *winOverlay) Update(g *Game) error {
	if w.copiedTTL > 0 {
		w.copiedTTL--
	}
	if w.handleKeys(g) {
		return nil
	}
	pos, pressed, _, _ := w.ptr.state()
	if !pressed {
		return nil
	}
	pts := []image.Point{pos}
	switch {
	case w.info.daily && w.btnShare.hit(pts):
		w.share()
	case !w.info.daily && w.btnMain.hit(pts):
		g.scene = newTableScene(w.info.g.Rules, newSeed())
	case w.btnMenu.hit(pts):
		g.scene = newMenuScene()
	}
	return nil
}

// handleKeys runs the primary action on Enter and menu on Esc; reports whether
// the scene changed.
func (w *winOverlay) handleKeys(g *Game) bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.scene = newMenuScene()
		return true
	}
	primary := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyN)
	if !primary {
		return false
	}
	if w.info.daily {
		w.share() // no new deal today — the primary action is to share
		return false
	}
	g.scene = newTableScene(w.info.g.Rules, newSeed())
	return true
}

func (w *winOverlay) share() {
	copyToClipboard(share.Text(share.Result{
		Variant:   w.info.g.Rules.Name(),
		Day:       w.info.day,
		Moves:     w.info.g.MoveCount,
		Undos:     w.info.undos,
		Hints:     w.info.hints,
		Streak:    w.info.streak,
		MaxStreak: w.info.maxStreak,
		URL:       w.info.shareURL,
	}))
	w.copiedTTL = 150
}

func (w *winOverlay) Draw(dst *ebiten.Image) {
	title := "YOU WON"
	drawText(dst, title, (W-textWidth(title, 6))/2, 120, colAccent, 6)

	line := fmt.Sprintf("%s — %d moves", w.info.g.Rules.Name(), w.info.g.MoveCount)
	drawText(dst, line, (W-textWidth(line, 2))/2, 220, colText, 2)

	drawText(dst, w.summary(), (W-textWidth(w.summary(), 1))/2, 260, colDim, 1)

	if w.info.daily {
		w.btnShare.draw(dst, true)
	} else {
		w.btnMain.draw(dst, true)
	}
	w.btnMenu.draw(dst, false)

	if w.copiedTTL > 0 {
		msg := "copied to clipboard!"
		drawText(dst, msg, (W-textWidth(msg, 1))/2, 300, colAmber, 1)
	}
}

// summary is the third line: daily streak, or the casual lifetime record.
func (w *winOverlay) summary() string {
	if w.info.daily {
		return fmt.Sprintf("Daily #%d · %d-day streak · best %d", w.info.day, w.info.streak, w.info.maxStreak)
	}
	st := stats.Get(string(w.info.g.Rules.ID()))
	return fmt.Sprintf("won %d of %d · best %d moves", st.Won, st.Played, st.BestMoves)
}
