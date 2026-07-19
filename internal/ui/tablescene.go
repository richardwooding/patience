package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/richardwooding/patience/internal/solitaire"
	"github.com/richardwooding/patience/internal/stats"
)

const dragThreshold = 4 // px of motion before a press becomes a drag

// tableScene is the game in progress: felt, piles, toolbar, drag machine.
type tableScene struct {
	g      *solitaire.Game
	layout tableLayout
	ptr    pointer

	// drag machine
	pressed       bool
	dragSrc       int
	dragIdx       int
	dragging      bool
	grabAt        image.Point // pointer offset within the grabbed card
	pressPos      image.Point
	dragPos       image.Point
	dropCandidate int

	// double-tap detection
	lastTapPile  int
	lastTapFrame int64
	frame        int64

	flights  []flight
	autoLeft int // frames until next auto-complete send; -1 idle
	toast    string
	toastTTL int
	cascade  *cascade

	btnNew, btnRestart, btnUndo, btnAuto, btnMenu button
}

func newTableScene(rules solitaire.Rules, seed uint64) *tableScene {
	ensureSprites()
	s := &tableScene{
		g:             solitaire.New(rules, seed),
		layout:        layoutFor(rules),
		autoLeft:      -1,
		dropCandidate: -1,
	}
	stats.Record(string(rules.ID()), stats.Dealt)
	const bw, bh, gap = 76, 24, 8
	x := 8
	place := func(b *button, label string) {
		*b = button{x: float32(x), y: 4, w: bw, h: bh, label: label}
		x += bw + gap
	}
	place(&s.btnNew, "new")
	place(&s.btnRestart, "restart")
	place(&s.btnUndo, "undo")
	place(&s.btnAuto, "auto")
	place(&s.btnMenu, "menu")
	return s
}

func (s *tableScene) Update(g *Game) error {
	s.frame++
	s.stepFlights()
	if s.cascade != nil {
		return s.updateCascade(g)
	}
	if s.toastTTL > 0 {
		s.toastTTL--
	}
	s.updateAutoComplete()

	if s.handleKeys(g) {
		return nil
	}

	pos, pressed, held, released := s.ptr.state()
	switch {
	case pressed:
		s.beginPress(pos)
	case held && s.pressed:
		s.continuePress(pos)
	case released && s.pressed:
		s.endPress(g, pos)
	}

	if s.g.Won() && s.cascade == nil {
		s.win()
	}
	return nil
}

func (s *tableScene) handleKeys(g *Game) bool {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyU):
		s.undo()
	case inpututil.IsKeyJustPressed(ebiten.KeyN):
		g.scene = newTableScene(s.g.Rules, newSeed())
		return true
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		g.scene = newTableScene(s.g.Rules, s.g.Seed)
		return true
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		s.startAutoComplete()
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		g.scene = newMenuScene()
		return true
	}
	return false
}

func (s *tableScene) undo() {
	s.autoLeft = -1
	if !s.g.Undo() {
		s.say("nothing to undo")
	}
}

func (s *tableScene) say(msg string) {
	s.toast = msg
	s.toastTTL = 150
}

// --- press / drag / drop ---

func (s *tableScene) beginPress(pos image.Point) {
	// toolbar first
	taps := []image.Point{pos}
	switch {
	case s.btnNew.hit(taps), s.btnRestart.hit(taps), s.btnUndo.hit(taps), s.btnAuto.hit(taps), s.btnMenu.hit(taps):
		// handled on release via endPress toolbar check; mark pressed
		s.pressed = true
		s.pressPos = pos
		s.dragSrc = -1
		return
	}
	pi, idx, ok := s.layout.hitCard(s.g, pos, s.ptr.hitInflate())
	if !ok {
		// empty stock slot still taps (Klondike recycle)
		if spi, sok := s.layout.hitPile(s.g, pos); sok && s.g.Piles[spi].Kind == solitaire.Stock {
			pi, idx = spi, -1
		} else {
			s.pressed = false
			return
		}
	}
	s.pressed = true
	s.pressPos = pos
	s.dragSrc = pi
	s.dragIdx = idx
	if idx >= 0 {
		r := s.layout.cardRect(s.g, pi, idx)
		s.grabAt = pos.Sub(r.Min)
	}
}

func (s *tableScene) continuePress(pos image.Point) {
	s.dragPos = pos
	if s.dragging {
		s.updateDropCandidate(pos)
		return
	}
	d := pos.Sub(s.pressPos)
	if d.X*d.X+d.Y*d.Y < dragThreshold*dragThreshold {
		return
	}
	// promote to drag if the grab is legal
	if s.dragSrc >= 0 && s.dragIdx >= 0 && s.g.Rules.CanPickUp(s.g, s.dragSrc, s.dragIdx) {
		s.dragging = true
		s.updateDropCandidate(pos)
	}
}

func (s *tableScene) updateDropCandidate(pos image.Point) {
	dragRect := image.Rectangle{Min: pos.Sub(s.grabAt)}
	dragRect.Max = dragRect.Min.Add(image.Pt(CardW, CardH))
	if pi, ok := s.layout.dropTarget(s.g, dragRect, s.dragSrc, s.dragIdx); ok {
		s.dropCandidate = pi
	} else {
		s.dropCandidate = -1
	}
}

func (s *tableScene) endPress(g *Game, pos image.Point) {
	defer func() {
		s.pressed = false
		s.dragging = false
		s.dropCandidate = -1
	}()

	if s.dragging {
		s.finishDrag()
		return
	}
	// tap: toolbar?
	taps := []image.Point{pos}
	switch {
	case s.btnNew.hit(taps):
		g.scene = newTableScene(s.g.Rules, newSeed())
		return
	case s.btnRestart.hit(taps):
		g.scene = newTableScene(s.g.Rules, s.g.Seed)
		return
	case s.btnUndo.hit(taps):
		s.undo()
		return
	case s.btnAuto.hit(taps):
		s.startAutoComplete()
		return
	case s.btnMenu.hit(taps):
		g.scene = newMenuScene()
		return
	}
	if s.dragSrc < 0 {
		return
	}
	s.tapCard()
}

func (s *tableScene) tapCard() {
	pile := &s.g.Piles[s.dragSrc]
	if pile.Kind == solitaire.Stock {
		if err := s.g.TapStock(); err != nil {
			s.say(err.Error())
		}
		return
	}
	// double-tap on a top card: send to foundation if legal
	if s.dragSrc == s.lastTapPile && s.frame-s.lastTapFrame < 21 {
		s.trySendToFoundation()
	}
	s.lastTapPile = s.dragSrc
	s.lastTapFrame = s.frame
}

func (s *tableScene) trySendToFoundation() {
	src := s.dragSrc
	p := &s.g.Piles[src]
	if len(p.Cards) == 0 {
		return
	}
	idx := len(p.Cards) - 1
	for di := range s.g.Piles {
		if s.g.Piles[di].Kind != solitaire.Foundation {
			continue
		}
		m := solitaire.Move{Src: src, Idx: idx, Dst: di}
		if s.g.Legal(m) {
			from := s.layout.positions(s.g, src)[idx]
			card := p.Cards[idx]
			if err := s.g.Apply(m); err == nil {
				s.addFlight(card, from, di, 8)
			}
			return
		}
	}
}

func (s *tableScene) finishDrag() {
	if s.dropCandidate >= 0 {
		m := solitaire.Move{Src: s.dragSrc, Idx: s.dragIdx, Dst: s.dropCandidate}
		if err := s.g.Apply(m); err == nil {
			return
		}
	}
	// snap back: flight from drop position to the card's resting place
	src := s.dragSrc
	if s.dragIdx < len(s.g.Piles[src].Cards) {
		card := s.g.Piles[src].Cards[s.dragIdx]
		to := s.layout.positions(s.g, src)[s.dragIdx]
		from := s.dragPos.Sub(s.grabAt)
		s.flights = append(s.flights, flight{
			sprite: sprite(card), from: from, to: to, dur: 8,
			pile: src, depth: len(s.g.Piles[src].Cards) - s.dragIdx,
		})
	}
}

// --- auto-complete ---

func (s *tableScene) startAutoComplete() {
	if !s.g.Rules.AutoCompleteReady(s.g) {
		s.say("not finishable yet")
		return
	}
	s.autoLeft = 1
}

func (s *tableScene) updateAutoComplete() {
	if s.autoLeft < 0 {
		return
	}
	s.autoLeft--
	if s.autoLeft > 0 {
		return
	}
	moves := s.g.Rules.SafeMoves(s.g)
	if len(moves) == 0 {
		s.autoLeft = -1
		return
	}
	m := moves[0]
	from := s.layout.positions(s.g, m.Src)[m.Idx]
	card := s.g.Piles[m.Src].Cards[m.Idx]
	if err := s.g.Apply(m); err != nil {
		s.autoLeft = -1
		return
	}
	s.addFlight(card, from, m.Dst, 6)
	s.autoLeft = 6
}

func (s *tableScene) addFlight(c solitaire.Card, from image.Point, dstPile int, dur float64) {
	to := s.layout.piles[dstPile].pos
	s.flights = append(s.flights, flight{
		sprite: sprite(c.WithFaceUp(true)), from: from, to: to, dur: dur,
		pile: dstPile, depth: 1,
	})
}

func (s *tableScene) stepFlights() {
	alive := s.flights[:0]
	for i := range s.flights {
		if !s.flights[i].step() {
			alive = append(alive, s.flights[i])
		}
	}
	s.flights = alive
}

func (s *tableScene) win() {
	stats.Record(string(s.g.Rules.ID()), stats.WonIn(s.g.MoveCount))
	s.cascade = newCascade(s.g, s.layout)
}

func (s *tableScene) updateCascade(g *Game) error {
	_, pressed, _, _ := s.ptr.state()
	if s.cascade.Update() || pressed || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.scene = newWinOverlay(s.g)
	}
	return nil
}

// --- draw ---

func (s *tableScene) Draw(dst *ebiten.Image) {
	if s.cascade != nil {
		s.cascade.Draw(dst)
		return
	}
	s.drawToolbar(dst)
	hidden := s.hiddenDepths()

	for pi := range s.g.Piles {
		s.drawPile(dst, pi, hidden[pi])
	}
	s.drawDrag(dst)
	for i := range s.flights {
		f := &s.flights[i]
		op := &ebiten.DrawImageOptions{}
		p := f.pos()
		op.GeoM.Translate(float64(p.X), float64(p.Y))
		dst.DrawImage(f.sprite, op)
	}
	if s.toastTTL > 0 {
		drawText(dst, s.toast, (W-textWidth(s.toast, 1))/2, H-22, colAmber, 1)
	}
}

// hiddenDepths counts, per pile, top cards suppressed by in-flight arrivals
// or an active drag.
func (s *tableScene) hiddenDepths() map[int]int {
	hidden := map[int]int{}
	for i := range s.flights {
		hidden[s.flights[i].pile] += s.flights[i].depth
	}
	if s.dragging {
		hidden[s.dragSrc] += len(s.g.Piles[s.dragSrc].Cards) - s.dragIdx
	}
	return hidden
}

func (s *tableScene) drawPile(dst *ebiten.Image, pi, hide int) {
	p := &s.g.Piles[pi]
	positions := s.layout.positions(s.g, pi)
	visible := len(p.Cards) - hide

	if visible <= 0 {
		slot := s.layout.piles[pi].pos
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(slot.X), float64(slot.Y))
		dst.DrawImage(slotSprite, op)
	}
	for i := 0; i < visible; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(positions[i].X), float64(positions[i].Y))
		dst.DrawImage(sprite(p.Cards[i]), op)
	}
	if s.dropCandidate == pi {
		r := s.layout.slotRect(pi)
		if n := len(p.Cards); n > 0 {
			r = s.layout.cardRect(s.g, pi, n-1)
		}
		accentOutline(dst, r)
	}
}

func (s *tableScene) drawDrag(dst *ebiten.Image) {
	if !s.dragging {
		return
	}
	base := s.dragPos.Sub(s.grabAt)
	cards := s.g.Piles[s.dragSrc].Cards[s.dragIdx:]
	for i, c := range cards {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(base.X)+3, float64(base.Y+i*fanUp)+3)
		op.ColorScale.ScaleAlpha(0.25) // drop shadow
		dst.DrawImage(sprite(c), op)
	}
	for i, c := range cards {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(base.X), float64(base.Y+i*fanUp))
		dst.DrawImage(sprite(c), op)
	}
}

func (s *tableScene) drawToolbar(dst *ebiten.Image) {
	s.btnNew.draw(dst, false)
	s.btnRestart.draw(dst, false)
	s.btnUndo.draw(dst, s.g.CanUndo())
	s.btnAuto.draw(dst, s.g.Rules.AutoCompleteReady(s.g))
	s.btnMenu.draw(dst, false)
	info := fmt.Sprintf("%s   moves %d", s.g.Rules.Name(), s.g.MoveCount)
	drawText(dst, info, W-16-textWidth(info, 1), 9, colDim, 1)
}
