//go:build js && wasm

package stats

import "syscall/js"

const key = "patience-stats"

func load() []byte {
	v := js.Global().Get("localStorage").Call("getItem", key)
	if v.IsNull() || v.IsUndefined() {
		return nil
	}
	return []byte(v.String())
}

func store(raw []byte) {
	defer func() { _ = recover() }() // private mode: storage may throw
	js.Global().Get("localStorage").Call("setItem", key, string(raw))
}
