//go:build !js

package ui

// autostartVariant is a no-op outside the browser.
func autostartVariant() string { return "" }
