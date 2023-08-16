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
package main

import (
	"context"
	"testing"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/controller"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestMain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup crosslink
	rootMpx := crosslink.NewMultiPlexer()
	cl := crosslink.NewCrosslink("crosslink", rootMpx)

	// setup colonio
	config := colonio.NewConfig()
	col, err := colonio.NewColonio(config)
	assert.NoError(t, err)
	err = col.Connect("https://localhost:8080/seed", "")
	assert.NoError(t, err)
	defer col.Disconnect()

	// run tests that can be run offline.
	suite.Run(t, cri.NewCriTest(cl))

	// test kvs
	suite.Run(t, kvs.NewAccountKvsTest(ctx, col))

	// test controller
	suite.Run(t, controller.NewAccountControllerTest(ctx, col))
	suite.Run(t, controller.NewNodeControllerTest())

	// test manager
}
