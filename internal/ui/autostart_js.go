//go:build js && wasm

package ui

import (
	"strconv"
	"strings"
	"syscall/js"
)

// autostartConfig reads the page URL to deep-link a deal: ?v=<variant-id>
// starts that variant, and an added &d=<day> starts that variant's daily deal
// for the given day number (how shared daily links are opened).
func autostartConfig() startConfig {
	var cfg startConfig
	search := js.Global().Get("location").Get("search").String()
	for q := range strings.SplitSeq(strings.TrimPrefix(search, "?"), "&") {
		if v, ok := strings.CutPrefix(q, "v="); ok {
			cfg.variant = v
		}
		if v, ok := strings.CutPrefix(q, "d="); ok {
			if day, err := strconv.Atoi(v); err == nil {
				cfg.day = day
				cfg.daily = true
			}
		}
	}
	return cfg
}
