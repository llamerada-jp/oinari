package crosslink

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"syscall/js"
)

type jsMessage struct {
	isServe bool
	id      uint32
	data    string // for serve
	tags    string // for serve
	reply   string // for reply
	message string // for reply
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
				impl.serve(msg.id, msg.data, msg.tags)
			} else {
				impl.replyFromJs(msg.id, msg.reply, msg.message)
			}
		}
	}(impl)

	impl.jsInstance.Set("serveToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		impl.jsChan <- jsMessage{
			isServe: true,
			id:      uint32(args[0].Int()),
			data:    args[1].String(),
			tags:    args[2].String(),
		}
		return nil
	}))

	impl.jsInstance.Set("callReplyToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		impl.jsChan <- jsMessage{
			isServe: false,
			id:      uint32(args[0].Int()),
			reply:   args[1].String(),
			message: args[2].String(),
		}
		return nil
	}))

	return impl
}

func (cl *crosslinkImpl) Call(path string, data []byte, tags map[string]string, cb func(result []byte, err error)) {
	tagsStr, err := json.Marshal(tags)
	if err != nil {
		log.Fatalln(err)
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

	cl.jsInstance.Call("callFromGo", js.ValueOf(id), js.ValueOf(path), js.ValueOf(string(data)), js.ValueOf(string(tagsStr)))
}

func (cl *crosslinkImpl) serve(id uint32, data, tagsStr string) {
	var tags map[string]string
	err := json.Unmarshal([]byte(tagsStr), &tags)
	if err != nil {
		log.Fatalln(err)
	}

	rw := &rwImpl{
		jsInstance: cl.jsInstance,
		id:         id,
	}

	cl.handler.Serve(data, tags, rw)
}

func (cl *crosslinkImpl) replyFromJs(id uint32, reply, message string) {
	cb, ok := cl.cbMap[id]
	if !ok {
		log.Fatalln("call back function is not exist")
	}
	defer delete(cl.cbMap, id)

	if message != "" {
		cb(nil, errors.New(message))
		return
	}
	cb([]byte(reply), nil)
}

func (rw *rwImpl) ReplySuccess(result string) {
	rw.jsInstance.Call("serveReplyFromGo", js.ValueOf(rw.id), js.ValueOf(result), js.ValueOf(""))
}

func (rw *rwImpl) ReplyError(message string) {
	rw.jsInstance.Call("serveReplyFromGo", js.ValueOf(rw.id), js.ValueOf(""), js.ValueOf(message))
}
