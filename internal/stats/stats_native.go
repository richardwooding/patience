//go:build !js

package stats

import (
	"os"
	"path/filepath"
)

func statsPath() (string, bool) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(dir, "patience", "stats.json"), true
}

func load() []byte {
	p, ok := statsPath()
	if !ok {
		return nil
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	return raw
}

func store(raw []byte) {
	p, ok := statsPath()
	if !ok {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, raw, 0o644)
}
