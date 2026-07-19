//go:build js && wasm

package ui

import (
	"strings"
	"syscall/js"
)

// autostartVariant reads ?v=<variant-id> from the page URL to deep-link a
// variant (e.g. ?v=freecell, ?v=spider-2).
func autostartVariant() string {
	search := js.Global().Get("location").Get("search").String()
	for q := range strings.SplitSeq(strings.TrimPrefix(search, "?"), "&") {
		if v, ok := strings.CutPrefix(q, "v="); ok {
			return v
		}
	}
	return ""
}
