package crosslink

import (
	"encoding/json"
	"errors"
	"math/rand"
	"syscall/js"
)

type ResponseWriter interface {
	ReplySuccess(result string)
	ReplyError(message string)
}

type Handler interface {
	Serve(data string, tags map[string]string, writer ResponseWriter)
}

type Crosslink interface {
	Call(data string, tags map[string]string, cb func(result string, err error))
}

type crosslinkImpl struct {
	jsInstance js.Value
	handler    Handler
	cbMap      map[uint32]func(string, error)
}

type rwImpl struct {
	jsInstance js.Value
	id         uint32
}

func NewCrosslink(jsName string, handler Handler) (Crosslink, error) {
	impl := &crosslinkImpl{
		jsInstance: js.Global().Get(jsName),
		handler:    handler,
	}

	impl.jsInstance.Set("serveToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		id := args[0].Int()
		data := args[1].String()
		tags := args[2].String()

		impl.serve(uint32(id), data, tags)

		return nil
	}))

	impl.jsInstance.Set("callReplyToGo", js.FuncOf(func(this js.Value, args []js.Value) any {
		id := args[0].Int()
		reply := args[1].String()
		message := args[2].String()

		impl.replyFromJs(uint32(id), reply, message)

		return nil
	}))

	return impl, nil
}

func (cl *crosslinkImpl) Call(data string, tags map[string]string, cb func(result string, err error)) {
	tagsStr, err := json.Marshal(tags)
	if err != nil {
		panic(err)
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

	cl.jsInstance.Call("callFromGo", js.ValueOf(id), js.ValueOf(data), js.ValueOf(tagsStr))
}

func (cl *crosslinkImpl) serve(id uint32, data, tagsStr string) {
	var tags map[string]string
	err := json.Unmarshal([]byte(data), &tags)
	if err != nil {
		panic(err)
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
		panic("call back function is not exist")
	}
	delete(cl.cbMap, id)

	if message != "" {
		cb("", errors.New(message))
	}
	cb(reply, nil)
}

func (rw *rwImpl) ReplySuccess(result string) {
	rw.jsInstance.Call("serveReplyFromGo", js.ValueOf(rw.id), js.ValueOf(result), js.ValueOf(""))
}

func (rw *rwImpl) ReplyError(message string) {
	rw.jsInstance.Call("serveReplyFromGo", js.ValueOf(rw.id), js.ValueOf(""), js.ValueOf(message))
}
