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

import "github.com/llamerada-jp/oinari/lib/crosslink"

const (
	ApplicationCrosslinkPath = "application/api/core"
	NodeCrosslinkPath        = "node/api/core"
)

type API interface {
	Name() string
	Setup(cl crosslink.Crosslink, apiMpx crosslink.MultiPlexer, errCh chan error) error
}

type Application interface {
	Setup(isInitialize bool, record []byte) error
	Marshal() ([]byte, error)
	Teardown(isFinalize bool) ([]byte, error)
}
