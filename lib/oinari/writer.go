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
package oinari

import (
	"encoding/json"
	"io"
	"log"

	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

var Writer io.Writer

type writer struct {
	cl   crosslink.Crosslink
	path string
}

func initWriter(cl crosslink.Crosslink, path string) error {
	Writer = &writer{
		cl:   cl,
		path: path,
	}

	return nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	type writeResponse struct {
		len int
		err error
	}
	resCh := make(chan writeResponse)

	w.cl.Call(w.path+"/output", core.OutputRequest{
		Payload: p,
	}, nil, func(b []byte, err error) {
		if err != nil {
			resCh <- writeResponse{
				len: 0,
				err: err,
			}
		}

		var res core.OutputResponse
		err = json.Unmarshal(b, &res)
		if err != nil {
			log.Fatalf("unmarshal response of output failed on oinari api: %s", err.Error())
		}

		resCh <- writeResponse{
			len: res.Length,
			err: nil,
		}
	})

	res := <-resCh
	return res.len, res.err
}
