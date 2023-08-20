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
	log.Fatal("TODO: implement UpdatePodAndNodeState() test")
	/// normal pattern
}
