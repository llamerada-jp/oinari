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
	"log"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/misc"
	"github.com/stretchr/testify/suite"
)

const ACCOUNT = "cat"
const NODE_ID = "012345678901234567890123456789ab"

type accountControllerTest struct {
	suite.Suite
	col        *misc.ColonioMock
	accountKvs kvs.AccountKvs
	impl       *accountControllerImpl
}

func NewAccountControllerTest() suite.TestingSuite {
	colonioMock := misc.NewColonioMock()
	accountKvs := kvs.NewAccountKvs(colonioMock)

	return &accountControllerTest{
		col:        colonioMock,
		accountKvs: accountKvs,
		impl: &accountControllerImpl{
			accountName: ACCOUNT,
			localNid:    NODE_ID,
			accountKvs:  accountKvs,
			logs:        make(map[string]*logEntry),
		},
	}
}

func (test *accountControllerTest) TestDealLocalResource() {
	account := "(=^_^=)"

	// dummy entry to avoid error when update
	test.col.KvsSet(api.GenerateAccountUuid(account), []byte("dummy"), 0)

	/// abnormal: return true if the data is invalid
	res, err := test.impl.DealLocalResource([]byte("invalid data"))
	test.Error(err)
	test.True(res)

	log.Fatal("TODO: implement DealLocalResource() test")

	/// normal: keep account.State.Pods before lifetime

	/// normal: delete account.State.Pods & the log entry after lifetime passed

	/// normal: delete pod log entry after 2 * lifetime passed

	/// normal: keep account.State.Nodes before lifetime
	/// normal: delete account.State.Nodes & the log entry after lifetime passed
	/// normal: delete node log entry after 2 * lifetime passed
	/// normal: return true if lifetime of account passed
}

func (test *accountControllerTest) TestGetAccountName() {
	test.Equal(ACCOUNT, test.impl.GetAccountName())
}

func (test *accountControllerTest) TestGetPodState() {
	podUuid1 := api.GeneratePodUuid()
	podUuid2 := api.GeneratePodUuid()

	test.accountKvs.Set(&api.Account{
		Meta: &api.ObjectMeta{
			Type:              api.ResourceTypeAccount,
			Name:              ACCOUNT,
			Owner:             ACCOUNT,
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              api.GenerateAccountUuid(ACCOUNT),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &api.AccountState{
			Pods: map[string]api.AccountPodState{
				podUuid1: {
					RunningNode: "012345678901234567890123456789ab",
					Timestamp:   "2023-04-15T17:30:40+09:00",
				},
				podUuid2: {
					RunningNode: "012345678901234567890123456789ac",
					Timestamp:   "2023-04-15T17:30:40+09:00",
				},
			},
			Nodes: make(map[string]api.AccountNodeState),
		},
	})

	/// normal pattern
	podStates, err := test.impl.GetPodState()
	test.NoError(err)
	test.Len(podStates, 2)
	test.Equal("012345678901234567890123456789ab", podStates[podUuid1].RunningNode)
	test.Equal("012345678901234567890123456789ac", podStates[podUuid2].RunningNode)

	/// normal: account data is not found
	test.col.DeleteKVSAll()
	podStates, err = test.impl.GetPodState()
	test.NoError(err)
	test.Len(podStates, 0)
}

func (test *accountControllerTest) TestGetNodeState() {
	nodeID1 := "012345678901234567890123456789ab"
	nodeID2 := "012345678901234567890123456789aa"

	test.accountKvs.Set(&api.Account{
		Meta: &api.ObjectMeta{
			Type:              api.ResourceTypeAccount,
			Name:              ACCOUNT,
			Owner:             ACCOUNT,
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              api.GenerateAccountUuid(ACCOUNT),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &api.AccountState{
			Pods: make(map[string]api.AccountPodState),
			Nodes: map[string]api.AccountNodeState{
				nodeID1: {
					Name:      "node1",
					Timestamp: "2023-04-15T17:30:40+09:00",
					NodeType:  api.NodeTypeServer,
				},
				nodeID2: {
					Name:      "node2",
					Timestamp: "2023-04-15T17:30:40+09:00",
					NodeType:  api.NodeTypePC,
				},
			},
		},
	})

	/// normal pattern
	nodeState, err := test.impl.GetNodeState()
	test.NoError(err)
	test.Len(nodeState, 2)
	test.Equal("node1", nodeState[nodeID1].Name)
	test.Equal(api.NodeTypeServer, nodeState[nodeID1].NodeType)
	test.Equal("node2", nodeState[nodeID2].Name)
	test.Equal(api.NodeTypePC, nodeState[nodeID2].NodeType)

	/// normal: account data is not found
	test.col.DeleteKVSAll()
	nodeState, err = test.impl.GetNodeState()
	test.NoError(err)
	test.Len(nodeState, 0)
}

func (test *accountControllerTest) TestUpdatePodAndNodeState() {
	test.col.DeleteKVSAll()
	account := "lucky"
	uuid1 := api.GeneratePodUuid()
	uuid2 := api.GeneratePodUuid()
	uuid3 := api.GeneratePodUuid()
	nodeID1 := "012345678901234567890123456789ab"
	nodeID2 := "012345678901234567890123456789ac"

	/// normal pattern: create new
	test.impl.UpdatePodAndNodeState(account,
		map[string]api.AccountPodState{
			uuid1: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:40+09:00",
			},
			uuid2: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:40+09:00",
			},
		},
		"012345678901234567890123456789ab",
		&api.AccountNodeState{
			Name:      "node name",
			Timestamp: "2023-04-15T17:30:40+09:00",
			NodeType:  api.NodeTypeServer,
			Latitude:  35.681167,
			Longitude: 139.767052,
			Altitude:  10.0,
		})

	data, err := test.impl.accountKvs.Get(account)
	test.NoError(err)
	test.Equal(account, data.Meta.Name)
	test.Equal(account, data.Meta.Owner)
	test.Equal(nodeID1, data.Meta.CreatorNode)
	test.Equal(api.GenerateAccountUuid(account), data.Meta.Uuid)
	test.Len(data.State.Pods, 2)
	test.Equal(nodeID1, data.State.Pods[uuid1].RunningNode)
	test.Equal("2023-04-15T17:30:40+09:00", data.State.Pods[uuid1].Timestamp)
	test.Equal(nodeID1, data.State.Pods[uuid2].RunningNode)
	test.Equal("2023-04-15T17:30:40+09:00", data.State.Pods[uuid2].Timestamp)
	test.Len(data.State.Nodes, 1)
	test.Equal("node name", data.State.Nodes[nodeID1].Name)
	test.Equal("2023-04-15T17:30:40+09:00", data.State.Nodes[nodeID1].Timestamp)
	test.Equal(api.NodeTypeServer, data.State.Nodes[nodeID1].NodeType)
	test.Equal(35.681167, data.State.Nodes[nodeID1].Latitude)
	test.Equal(139.767052, data.State.Nodes[nodeID1].Longitude)
	test.Equal(10.0, data.State.Nodes[nodeID1].Altitude)

	/// normal pattern: update exists
	test.impl.UpdatePodAndNodeState(account,
		map[string]api.AccountPodState{
			uuid2: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:41+09:00",
			},
			uuid3: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:41+09:00",
			},
		},
		nodeID1,
		&api.AccountNodeState{
			Name:      "node name",
			Timestamp: "2023-04-15T17:30:41+09:00",
			NodeType:  api.NodeTypeServer,
			Latitude:  36.681167,
			Longitude: 140.767052,
			Altitude:  11.0,
		})
	data, err = test.impl.accountKvs.Get(account)
	test.NoError(err)
	test.Len(data.State.Pods, 3)
	test.Equal(nodeID1, data.State.Pods[uuid1].RunningNode)
	test.Equal("2023-04-15T17:30:40+09:00", data.State.Pods[uuid1].Timestamp)
	test.Equal(nodeID1, data.State.Pods[uuid2].RunningNode)
	test.Equal("2023-04-15T17:30:41+09:00", data.State.Pods[uuid2].Timestamp)
	test.Equal(nodeID1, data.State.Pods[uuid3].RunningNode)
	test.Equal("2023-04-15T17:30:41+09:00", data.State.Pods[uuid3].Timestamp)
	test.Len(data.State.Nodes, 1)
	test.Equal("node name", data.State.Nodes[nodeID1].Name)

	// normal pattern: update pod state only
	test.impl.UpdatePodAndNodeState(account,
		map[string]api.AccountPodState{
			uuid2: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:42+09:00",
			},
			uuid3: {
				RunningNode: nodeID1,
				Timestamp:   "2023-04-15T17:30:42+09:00",
			},
		}, "", nil)
	data, err = test.impl.accountKvs.Get(account)
	test.NoError(err)
	test.Len(data.State.Pods, 3)
	test.Equal(nodeID1, data.State.Pods[uuid2].RunningNode)
	test.Equal("2023-04-15T17:30:42+09:00", data.State.Pods[uuid2].Timestamp)
	test.Equal(nodeID1, data.State.Pods[uuid3].RunningNode)
	test.Equal("2023-04-15T17:30:42+09:00", data.State.Pods[uuid3].Timestamp)

	// normal pattern: update node state only
	test.impl.UpdatePodAndNodeState(account,
		nil, nodeID2,
		&api.AccountNodeState{
			Name:      "node name 2",
			Timestamp: "2023-04-15T17:30:41+09:00",
			NodeType:  api.NodeTypeGrass,
			Latitude:  37.681167,
			Longitude: 141.767052,
			Altitude:  12.0,
		})
	data, err = test.impl.accountKvs.Get(account)
	test.NoError(err)
	test.Len(data.State.Pods, 3)
	test.Len(data.State.Nodes, 2)
	test.Equal("node name", data.State.Nodes[nodeID1].Name)
	test.Equal(api.NodeTypeServer, data.State.Nodes[nodeID1].NodeType)
	test.Equal(36.681167, data.State.Nodes[nodeID1].Latitude)
	test.Equal(140.767052, data.State.Nodes[nodeID1].Longitude)
	test.Equal(11.0, data.State.Nodes[nodeID1].Altitude)
	test.Equal("node name 2", data.State.Nodes[nodeID2].Name)
	test.Equal(api.NodeTypeGrass, data.State.Nodes[nodeID2].NodeType)
	test.Equal(37.681167, data.State.Nodes[nodeID2].Latitude)
	test.Equal(141.767052, data.State.Nodes[nodeID2].Longitude)
	test.Equal(12.0, data.State.Nodes[nodeID2].Altitude)
}
