/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"
	"time"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/agent/core"
	"github.com/llamerada-jp/oinari/agent/crosslink"
	"github.com/llamerada-jp/oinari/agent/global"
	"github.com/llamerada-jp/oinari/agent/local"
)

const (
	gltfLoaderName = "gltfLoader"
	fieldBinder    = "newField"
)

type asset struct {
	jsAssert js.Value
	err      error
}

type field struct {
	jsField js.Value
}

type object struct {
	asset    *asset
	field    *field
	jsObject js.Value
}

func newGltfAssert(url string) *asset {
	a := &asset{
		jsAssert: js.Null(),
	}

	js.Global().Call(gltfLoaderName, url,
		js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			// onLoad
			a.jsAssert = args[0]
			return nil
		}),
		js.Null(), // onProgress
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// onError
			a.err = fmt.Errorf(args[0].String())
			return nil
		}))

	return a
}

func (a *asset) hasError() error {
	return a.err
}

func (a *asset) checkLoaded() bool {
	return !a.jsAssert.IsNull()
}

func newField() *field {
	return &field{
		jsField: js.Global().Call(fieldBinder),
	}
}

func newObject(asset *asset, field *field) *object {
	return &object{
		asset:    asset,
		field:    field,
		jsObject: js.Null(),
	}
}

func (o *object) tryBind() bool {
	// Already binded the object.
	if !o.jsObject.IsNull() {
		return true
	}

	// The asset has not loaded yet.
	if !o.asset.checkLoaded() {
		return false
	}

	o.jsObject = o.asset.jsAssert.Call("clone", js.ValueOf(true))
	o.field.jsField.Call("add", o.jsObject)

	return true
}

func initCrosslink() (crosslink.Crosslink, crosslink.MultiPlexer) {
	mpxRoot := crosslink.NewMultiPlexer()
	cl := crosslink.NewCrosslink("crosslink", mpxRoot)

	return cl, mpxRoot
}

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx, _ := context.WithCancel(context.Background())

	cl, mpxRoot := initCrosslink()
	config := colonio.NewConfig()
	col, err := colonio.NewColonio(config)
	if err != nil {
		log.Fatalln(err)
	}

	gcd := global.NewCommandDriver(col)
	lcd := local.NewCommandDriver(cl)
	seh := newSystemEventHandler(col)

	sys := core.NewSystem(col, seh, gcd, lcd)
	go func() {
		err := sys.Start(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	local.InitHandler(sys, mpxRoot)

	// send a message that tell initialization complete
	cl.Call("", map[string]string{
		crosslink.TAG_PATH: "system/initializationComplete",
	}, func(result string, err error) {
		if err != nil {
			log.Fatalln(err)
		}
	})

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
		}
	}
}
