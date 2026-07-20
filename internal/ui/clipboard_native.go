//go:build !js

package ui

import "fmt"

// copyToClipboard has no system clipboard to reach natively, so it prints the
// blurb for the player to copy from the terminal.
func copyToClipboard(s string) {
	fmt.Println("\n--- share ---\n" + s + "\n-------------")
}
