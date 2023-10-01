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
package crosslink

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"syscall/js"
)

type jsMessage struct {
	isServe     bool
	id          uint32
	dataRaw     []byte // for serve
	tagsRaw     []byte // for serve
	responseRaw []byte // for response
	message     string // for response
}

type crosslinkImpl struct {
	jsInstance js.Value
	handler    Handler
	cbMap      map[uint32]func([]byte, error)
	jsChan     chan jsMessage
}

type rwImpl struct {
	jsInstance js.Value
	id         uint32
}

func NewCrosslink(jsName string, handler Handler) Crosslink {
	impl := &crosslinkImpl{
		jsInstance: js.Global().Get(jsName),
		handler:    handler,
		cbMap:      make(map[uint32]func([]byte, error)),
		jsChan:     make(chan jsMessage, 10),
	}

	// exec serve and replyFromJs method on a go routine to avoid blocking js thread
	go func(impl *crosslinkImpl) {
		for msg := range impl.jsChan {
			if msg.isServe {
				impl.serve(msg.id, msg.dataRaw, msg.tagsRaw)
			} else {
				impl.replyFromJs(msg.id, msg.responseRaw, msg.message)
			}
		}
	}(impl)

	impl.jsInstance.Set("serveToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		impl.jsChan <- jsMessage{
			isServe: true,
			id:      uint32(args[0].Int()),
			dataRaw: []byte(args[1].String()),
			tagsRaw: []byte(args[2].String()),
		}
		return nil
	}))

	impl.jsInstance.Set("replyToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		impl.jsChan <- jsMessage{
			isServe:     false,
			id:          uint32(args[0].Int()),
			responseRaw: []byte(args[1].String()),
			message:     args[2].String(),
		}
		return nil
	}))

	return impl
}

func (cl *crosslinkImpl) Call(path string, obj any, tags map[string]string, cb func([]byte, error)) {
	objStr := ""
	if obj != nil {
		objBin, err := json.Marshal(obj)
		if err != nil {
			log.Fatalf("marshaling a parameter failed on crosslink call method: %s", err.Error())
		}
		objStr = string(objBin)
	}

	tagsStr := ""
	if tags != nil {
		tagsBin, err := json.Marshal(tags)
		if err != nil {
			log.Fatalf("marshaling the tag failed on crosslink call method: %s", err.Error())
		}
		tagsStr = string(tagsBin)
	}

	var id uint32
	for {
		id = rand.Uint32()
		_, ok := cl.cbMap[id]
		if !ok {
			break
		}
	}
	cl.cbMap[id] = cb

	cl.jsInstance.Call("callFromGo", js.ValueOf(id), js.ValueOf(path), js.ValueOf(objStr), js.ValueOf(tagsStr))
}

func (cl *crosslinkImpl) serve(id uint32, dtaRaw, tagRaw []byte) {
	var tags map[string]string
	err := json.Unmarshal(tagRaw, &tags)
	if err != nil {
		log.Fatalf("unmarshal the tag failed on crosslink serve method: %s", err.Error())
	}

	rw := &rwImpl{
		jsInstance: cl.jsInstance,
		id:         id,
	}

	go cl.handler.Serve(dtaRaw, tags, rw)
}

func (cl *crosslinkImpl) replyFromJs(id uint32, responseRaw []byte, message string) {
	cb, ok := cl.cbMap[id]
	if !ok {
		log.Fatalln("call back function is not exist on crosslink")
	}
	defer delete(cl.cbMap, id)

	if message != "" {
		go cb(nil, errors.New(message))
		return
	}
	go cb(responseRaw, nil)
}

func (rw *rwImpl) ReplySuccess(response any) {
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("marshaling the response failed on crosslink ReplySuccess: %s", err.Error())
	}
	rw.jsInstance.Call("replyFromGo", js.ValueOf(rw.id), js.ValueOf(string(responseJson)), js.ValueOf(""))
}

func (rw *rwImpl) ReplyError(message string) {
	rw.jsInstance.Call("replyFromGo", js.ValueOf(rw.id), js.ValueOf(""), js.ValueOf(message))
}
