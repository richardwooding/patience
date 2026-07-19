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

func NewGame() *Game {
	if id := autostartVariant(); id != "" {
		for _, v := range solitaire.Variants() {
			if string(v.ID()) == id {
				return &Game{scene: newTableScene(v, newSeed())}
			}
		}
	}
	return &Game{scene: newMenuScene()}
}

func (g *Game) Update() error              { return g.scene.Update(g) }
func (g *Game) Draw(dst *ebiten.Image)     { dst.Fill(colBG); g.scene.Draw(dst) }
func (g *Game) Layout(_, _ int) (int, int) { return W, H }
