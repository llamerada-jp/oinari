package crosslink

import (
	"encoding/json"
	"fmt"
	"log"
)

type funcHandlerImpl struct {
	f func(dataRaw []byte, tags map[string]string, writer ResponseWriter)
}

func NewFuncHandler[T any](f func(param *T, tags map[string]string, writer ResponseWriter)) Handler {
	return &funcHandlerImpl{
		f: func(dataRaw []byte, tags map[string]string, writer ResponseWriter) {
			var t T
			err := json.Unmarshal(dataRaw, &t)
			if err != nil {
				writer.ReplyError(fmt.Sprintf("json unmarshal error, %v %s", err, string(dataRaw)))
				return
			}

			f(&t, tags, writer)
		},
	}
}

func (f *funcHandlerImpl) Serve(dataRaw []byte, tags map[string]string, writer ResponseWriter) {
	if kind, ok := tags[TAG_PATH_MATCH_KIND]; ok {
		if kind != PATH_MATCH_KIND_EXACT {
			log.Fatalln("crosslink func handler should be called with exact match path")
		}
	}
	f.f(dataRaw, tags, writer)
}
