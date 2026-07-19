package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// pointer unifies mouse and single-touch into pressed/held/released events
// with a current position — the drag machine's input.
type pointer struct {
	touchID ebiten.TouchID
	byTouch bool
	active  bool
}

// state returns (pos, justPressed, held, justReleased) for this frame.
func (pt *pointer) state() (image.Point, bool, bool, bool) {
	// touch first: it wins over the synthetic mouse position on mobile
	if ids := inpututil.AppendJustPressedTouchIDs(nil); len(ids) > 0 {
		pt.touchID = ids[0]
		pt.byTouch = true
		pt.active = true
		x, y := ebiten.TouchPosition(pt.touchID)
		return image.Pt(x, y), true, true, false
	}
	if pt.byTouch && pt.active {
		if inpututil.IsTouchJustReleased(pt.touchID) {
			pt.active = false
			x, y := inpututil.TouchPositionInPreviousTick(pt.touchID)
			return image.Pt(x, y), false, false, true
		}
		x, y := ebiten.TouchPosition(pt.touchID)
		return image.Pt(x, y), false, true, false
	}

	x, y := ebiten.CursorPosition()
	pos := image.Pt(x, y)
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		pt.byTouch = false
		pt.active = true
		return pos, true, true, false
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && pt.active && !pt.byTouch:
		pt.active = false
		return pos, false, false, true
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && pt.active && !pt.byTouch:
		return pos, false, true, false
	}
	return pos, false, false, false
}

// hitInflate is the touch-target inflation in logical pixels.
func (pt *pointer) hitInflate() int {
	if pt.byTouch {
		return 6
	}
	return 0
}
