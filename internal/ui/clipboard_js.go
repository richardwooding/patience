//go:build js && wasm

package ui

import "syscall/js"

// copyToClipboard writes s to the system clipboard via the async Clipboard
// API. Called from a click/keypress handler, so it has the transient user
// activation the API requires.
func copyToClipboard(s string) {
	nav := js.Global().Get("navigator")
	if cb := nav.Get("clipboard"); cb.Truthy() {
		cb.Call("writeText", s)
	}
}
