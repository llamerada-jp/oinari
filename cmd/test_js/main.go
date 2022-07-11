//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"fmt"
	"syscall/js"
)

const instanceName = "goTest"

var jsInstance js.Value

func setup(ctx context.Context) {
	fmt.Println("setup")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	jsInstance = js.Global().Get(instanceName)
	jsInstance.Set("stop", js.FuncOf(func(this js.Value, args []js.Value) any {
		cancel()
		return nil
	}))

	setup(ctx)
	for {
		fmt.Println("loop")
		select {
		case <-ctx.Done():
			return
		}
	}
}
