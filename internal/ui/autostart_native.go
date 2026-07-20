//go:build !js

package ui

// autostartConfig is a no-op outside the browser.
func autostartConfig() startConfig { return startConfig{} }
