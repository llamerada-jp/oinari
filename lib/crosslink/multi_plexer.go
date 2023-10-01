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
	"fmt"
	"log"
	"regexp"
)

type mpxImpl struct {
	defaultHandler Handler
	handlers       map[string]Handler
}

func NewMultiPlexer() MultiPlexer {
	return &mpxImpl{
		defaultHandler: NewFuncHandler(func(_ *interface{}, tags map[string]string, writer ResponseWriter) {
			writer.ReplyError(fmt.Sprintf("handler for %s is not defined", tags[TAG_PATH]))
		}),
		handlers: make(map[string]Handler),
	}
}

func (m *mpxImpl) Serve(dataRaw []byte, tags map[string]string, writer ResponseWriter) {
	multiPlexerSpliter := regexp.MustCompile(`^/?([^/]*)/?(.*)$`)

	var path string
	var leaf string
	var ok bool

	if path, ok = tags[TAG_PATH]; !ok {
		log.Fatalln("`path` tag should be set in crosslink multi plexer")
	}

	if leaf, ok = tags[TAG_LEAF]; !ok {
		leaf = path
	}

	r := multiPlexerSpliter.FindStringSubmatch(leaf)
	dir := ""
	newLeaf := ""
	if r != nil {
		dir = r[1]
		newLeaf = r[2]
	}

	newTags := make(map[string]string)
	for k, v := range tags {
		newTags[k] = v
	}
	if len(newLeaf) == 0 {
		newTags[TAG_PATH_MATCH_KIND] = PATH_MATCH_KIND_EXACT
		delete(newTags, TAG_LEAF)
	} else {
		newTags[TAG_PATH_MATCH_KIND] = PATH_MATCH_KIND_HEAD
		newTags[TAG_LEAF] = newLeaf
	}

	if len(dir) != 0 {
		if handler, ok := m.handlers[dir]; ok {
			handler.Serve(dataRaw, newTags, writer)
			return
		}
	}

	delete(newTags, TAG_PATH_MATCH_KIND)
	delete(newTags, TAG_LEAF)
	m.defaultHandler.Serve(dataRaw, newTags, writer)
}

func (m *mpxImpl) SetHandler(pattern string, handler Handler) {
	m.handlers[pattern] = handler
}

func (m *mpxImpl) SetDefaultHandler(handler Handler) {
	m.defaultHandler = handler
}
