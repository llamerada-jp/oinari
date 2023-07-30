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
package driver

import (
	"log"

	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type FrontendDriver interface {
	// send a message that tell initialization complete
	TellInitComplete() error
}

type frontendDriverImpl struct {
	cl crosslink.Crosslink
}

func NewFrontendDriver(cl crosslink.Crosslink) FrontendDriver {
	return &frontendDriverImpl{
		cl: cl,
	}
}

func (impl *frontendDriverImpl) TellInitComplete() error {
	impl.cl.Call("frontend/onInitComplete", nil, nil,
		func(_ []byte, err error) {
			if err != nil {
				log.Fatalln(err)
			}
		})
	return nil
}
