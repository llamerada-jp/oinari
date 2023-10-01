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
