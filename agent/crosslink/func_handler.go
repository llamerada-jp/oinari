package crosslink

import (
	"encoding/json"
	"fmt"
	"log"
)

type funcHandlerImpl struct {
	f func(data string, tags map[string]string, writer ResponseWriter)
}

type objWriterImpl struct {
	writer ResponseWriter
}

func NewFuncHandler(f func(data string, tags map[string]string, writer ResponseWriter)) Handler {
	return &funcHandlerImpl{
		f: f,
	}
}

func NewFuncObjHandler[T any](f func(param *T, tags map[string]string, writer ResponseObjWriter)) Handler {
	return &funcHandlerImpl{
		f: func(dataStr string, tags map[string]string, writer ResponseWriter) {
			var t T
			err := json.Unmarshal([]byte(dataStr), &t)
			if err != nil {
				fmt.Printf("json error:%s", dataStr)
				writer.ReplyError("json unmarshal error")
				return
			}

			f(&t, tags, &objWriterImpl{
				writer: writer,
			})
		},
	}
}

func (f *funcHandlerImpl) Serve(data string, tags map[string]string, writer ResponseWriter) {
	if kind, ok := tags[TAG_PATH_MATCH_KIND]; ok {
		if kind != PATH_MATCH_KIND_EXACT {
			log.Fatalln("func handler should be called with exact match path")
		}
	}
	f.f(data, tags, writer)
}

func (w *objWriterImpl) ReplySuccess(result any) {
	if result == nil {
		w.writer.ReplySuccess("")
		return
	}
	res, err := json.Marshal(result)
	if err != nil {
		log.Fatalln(err)
	}
	w.writer.ReplySuccess(string(res))
}

func (w *objWriterImpl) ReplyError(message string) {
	w.writer.ReplyError(message)
}
