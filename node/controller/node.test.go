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
package controller

import (
	"github.com/stretchr/testify/suite"
)

type nodeControllerTest struct {
	suite.Suite
	impl *nodeControllerImpl
}

const (
	nodeID = "test-nid"
)

func NewNodeControllerTest() suite.TestingSuite {
	return &nodeControllerTest{
		impl: &nodeControllerImpl{
			nodeID: nodeID,
		},
	}
}

func (test *nodeControllerTest) TestNid() {
	test.Equal(nodeID, test.impl.GetNid())
}

// TODO add tests
