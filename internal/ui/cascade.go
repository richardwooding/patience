package ui

import (
	"image"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/richardwooding/patience/internal/solitaire"
)

// cascade is the classic win animation: cards launch from the foundations
// one at a time, bounce along the bottom edge, and leave trails on a
// persistent offscreen image that is never cleared (chipdeck phosphor-layer
// pattern — trails cost nothing).
type cascade struct {
	trail   *ebiten.Image
	queue   []solitaire.Card
	origins []image.Point
	active  []bouncer
	spawnIn int
	rng     *rand.Rand
}

type bouncer struct {
	sprite *ebiten.Image
	x, y   float64
	vx, vy float64
}

func newCascade(g *solitaire.Game, tl tableLayout) *cascade {
	c := &cascade{
		trail: ebiten.NewImage(W, H),
		rng:   rand.New(rand.NewPCG(g.Seed, uint64(g.MoveCount))),
	}
	c.trail.Fill(colBG)
	// launch foundation cards from the top down
	for pi := range g.Piles {
		if g.Piles[pi].Kind != solitaire.Foundation {
			continue
		}
		for i := len(g.Piles[pi].Cards) - 1; i >= 0; i-- {
			c.queue = append(c.queue, g.Piles[pi].Cards[i].WithFaceUp(true))
			c.origins = append(c.origins, tl.piles[pi].pos)
		}
	}
	return c
}

// Update advances the cascade; true when every card has left the screen.
func (c *cascade) Update() bool {
	if c.spawnIn <= 0 && len(c.queue) > 0 {
		card, origin := c.queue[0], c.origins[0]
		c.queue, c.origins = c.queue[1:], c.origins[1:]
		vx := 1 + c.rng.Float64()*3
		if c.rng.IntN(2) == 0 {
			vx = -vx
		}
		c.active = append(c.active, bouncer{
			sprite: sprite(card),
			x:      float64(origin.X), y: float64(origin.Y),
			vx: vx, vy: -c.rng.Float64() * 6,
		})
		c.spawnIn = 8
	}
	c.spawnIn--

	alive := c.active[:0]
	for _, b := range c.active {
		b.x += b.vx
		b.y += b.vy
		b.vy += 0.4
		if b.y > H-CardH {
			b.y = H - CardH
			b.vy *= -0.72
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.x, b.y)
		c.trail.DrawImage(b.sprite, op)
		if b.x > -CardW && b.x < W {
			alive = append(alive, b)
		}
	}
	c.active = alive
	return len(c.queue) == 0 && len(c.active) == 0
}

func (c *cascade) Draw(dst *ebiten.Image) {
	dst.DrawImage(c.trail, nil)
}
