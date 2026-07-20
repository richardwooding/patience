// Package ui is the Ebitengine front end: menu, table with the drag
// machine, win cascade and overlay.
package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/richardwooding/patience/internal/solitaire"
)

type scene interface {
	Update(g *Game) error
	Draw(dst *ebiten.Image)
}

// Game is the ebiten.Game holding the current scene.
type Game struct {
	scene scene
}

// startConfig is a parsed deep-link from the page URL.
type startConfig struct {
	variant string
	day     int
	daily   bool
}

// playBase is the deployed play page, used to build shareable daily links.
const playBase = "https://richardwooding.github.io/patience/play/"

func NewGame() *Game {
	cfg := autostartConfig()
	if cfg.variant != "" {
		for _, v := range solitaire.Variants() {
			if string(v.ID()) != cfg.variant {
				continue
			}
			if cfg.daily {
				return &Game{scene: newDailyScene(v, cfg.day)}
			}
			return &Game{scene: newTableScene(v, newSeed())}
		}
	}
	return &Game{scene: newMenuScene()}
}

func (g *Game) Update() error              { return g.scene.Update(g) }
func (g *Game) Draw(dst *ebiten.Image)     { dst.Fill(colBG); g.scene.Draw(dst) }
func (g *Game) Layout(_, _ int) (int, int) { return W, H }
