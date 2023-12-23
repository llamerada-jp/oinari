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

const (
	TAG_PATH            = "path"
	TAG_LEAF            = "leaf"
	TAG_PATH_MATCH_KIND = "matchKind"

	PATH_MATCH_KIND_EXACT = "E"
	PATH_MATCH_KIND_HEAD  = "H"
)

type ResponseWriter interface {
	ReplySuccess(response any)
	ReplyError(message string)
}

type Handler interface {
	Serve(dataRaw []byte, tags map[string]string, writer ResponseWriter)
}

type Crosslink interface {
	Call(path string, obj any, tags map[string]string, cb func([]byte, error))
}

type MultiPlexer interface {
	Handler
	SetHandler(pattern string, handler Handler)
	SetDefaultHandler(handler Handler)
}
