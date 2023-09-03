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
	"math"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/mock"
	"github.com/stretchr/testify/suite"
)

type nodeControllerTest struct {
	suite.Suite
	col  *mock.Colonio
	impl *nodeControllerImpl
}

const (
	nodeID   = "test-nid"
	nodeName = "test-node"
	nodeType = api.NodeTypeMobile
)

func NewNodeControllerTest() suite.TestingSuite {
	colonioMock := mock.NewColonioMock()

	return &nodeControllerTest{
		col: colonioMock,
		impl: &nodeControllerImpl{
			col:       colonioMock,
			nodeID:    nodeID,
			nodeName:  nodeName,
			nodeType:  nodeType,
			latitude:  12.345,
			longitude: 67.890,
			altitude:  10.0,
		},
	}
}

func (test *nodeControllerTest) TestGetNid() {
	test.Equal(nodeID, test.impl.GetNid())
}

func (test *nodeControllerTest) TestGetNodeState() {
	nodeState := test.impl.GetNodeState()
	test.Equal(nodeName, nodeState.Name)
	test.Empty(nodeState.Timestamp)
	test.Equal(nodeType, nodeState.NodeType)
	test.InDelta(12.345, nodeState.Latitude, 0.0001)
	test.InDelta(67.890, nodeState.Longitude, 0.0001)
	test.InDelta(10.0, nodeState.Altitude, 0.0001)
}

func (test *nodeControllerTest) TestSetPosition() {
	test.NoError(test.impl.SetPosition(34.345, 88.890, 11.0))
	test.InDelta(88.890*math.Pi/180, test.col.PositionX, 0.0001)
	test.InDelta(34.345*math.Pi/180, test.col.PositionY, 0.0001)

	nodeState := test.impl.GetNodeState()
	test.InDelta(34.345, nodeState.Latitude, 0.0001)
	test.InDelta(88.890, nodeState.Longitude, 0.0001)
	test.InDelta(11.0, nodeState.Altitude, 0.0001)
}

// TODO add tests
