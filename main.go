// patience is retro solitaire — Klondike, FreeCell, and Spider — with a
// pure-Go rules core under an Ebitengine front end. One codebase runs
// native for development and as WebAssembly in the browser.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/richardwooding/patience/internal/ui"
)

func main() {
	ebiten.SetWindowSize(ui.W, ui.H)
	ebiten.SetWindowTitle("patience")
	if err := ebiten.RunGame(ui.NewGame()); err != nil {
		log.Fatal(err)
	}
}
