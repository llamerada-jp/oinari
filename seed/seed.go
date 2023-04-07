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
package seed

import (
	"net/http"

	colonioSeed "github.com/llamerada-jp/colonio/go/seed"
)

func InitSeed(mux *http.ServeMux, documentRoot string, seedConfig *colonioSeed.Config) error {
	// Set HTTP headers to enable SharedArrayBuffer for WebAssembly Threads
	headers := make(map[string]string)
	headers["Cross-Origin-Embedder-Policy"] = "require-corp"
	headers["Cross-Origin-Opener-Policy"] = "same-origin"
	headers["Cross-Origin-Resource-Policy"] = "cross-origin"

	// Publish static documents using HTTP(S).
	mux.Handle("/", &headerWrapper{
		handler: http.FileServer(http.Dir(documentRoot)),
		headers: headers,
	})

	seed, err := colonioSeed.NewSeed(seedConfig)
	if err != nil {
		return err
	}
	mux.Handle("/seed", &headerWrapper{
		handler: seed,
		headers: headers,
	})
	if err := seed.Start(); err != nil {
		return err
	}

	return nil
}

type headerWrapper struct {
	handler http.Handler
	headers map[string]string
}

func (h *headerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headerWriter := w.Header()
	for k, v := range h.headers {
		headerWriter.Add(k, v)
	}
	h.handler.ServeHTTP(w, r)
}
