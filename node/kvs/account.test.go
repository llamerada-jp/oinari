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
package kvs

import (
	"encoding/json"
	"math"

	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/node/mock"
	"github.com/stretchr/testify/suite"
)

type accountKvsTest struct {
	suite.Suite
	col  *mock.Colonio
	impl *accountKvsImpl
}

func NewAccountKvsTest() suite.TestingSuite {
	colonioMock := mock.NewColonioMock()
	return &accountKvsTest{
		col: colonioMock,
		impl: &accountKvsImpl{
			col: colonioMock,
		},
	}
}

func (test *accountKvsTest) TestGet() {
	// return nil if the record is not exist
	accountName := "get-not-exist"
	key := test.impl.getKey(accountName)
	_, err := test.col.KvsGet(key)
	test.Error(err)
	accountGet, err := test.impl.Get(accountName)
	test.NoError(err)
	test.Nil(accountGet)

	// return nil if the record is nil
	accountName = "get-account-nil"
	key = test.impl.getKey(accountName)
	err = test.col.KvsSet(key, nil, 0)
	test.NoError(err)
	record, err := test.col.KvsGet(key)
	test.NoError(err)
	test.True(record.IsNil())
	accountGet, err = test.impl.Get(accountName)
	test.NoError(err)
	test.Nil(accountGet)

	// can get valid account record
	accountName = "get-account-valid"
	key = test.impl.getKey(accountName)
	podUuid := core.GeneratePodUuid()
	accountSet := &core.Account{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypeAccount,
			Name:              accountName,
			Owner:             accountName,
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              core.GenerateAccountUuid(accountName),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &core.AccountState{
			Pods: map[string]core.AccountPodState{
				podUuid: {
					RunningNode: "012345678901234567890123456789ab",
					Timestamp:   "2023-04-15T17:30:40+09:00",
				},
			},
			Nodes: map[string]core.AccountNodeState{
				"012345678901234567890123456789ab": {
					Name:      "test-node",
					NodeType:  core.NodeTypeMobile,
					Timestamp: "2023-04-15T17:30:40+09:00",
				},
			},
		},
	}

	test.NoError(accountSet.Validate())
	raw, err := json.Marshal(accountSet)
	test.NoError(err)

	err = test.col.KvsSet(key, raw, 0)
	test.NoError(err)
	accountGet, err = test.impl.Get(accountName)
	test.NoError(err)
	test.NotNil(accountGet)
	test.Equal(core.ResourceTypeAccount, accountGet.Meta.Type)
	test.Equal(accountName, accountGet.Meta.Name)
	test.Equal(accountName, accountGet.Meta.Owner)
	test.Equal("012345678901234567890123456789ab", accountGet.Meta.CreatorNode)
	test.Equal(core.GenerateAccountUuid(accountName), accountGet.Meta.Uuid)
	test.Equal("2023-04-15T17:30:40+09:00", accountGet.Meta.DeletionTimestamp)
	test.Len(accountGet.State.Pods, 1)
	test.Equal("012345678901234567890123456789ab", accountGet.State.Pods[podUuid].RunningNode)
	test.Equal("2023-04-15T17:30:40+09:00", accountGet.State.Pods[podUuid].Timestamp)
	test.Len(accountGet.State.Nodes, 1)
	test.Equal("test-node", accountGet.State.Nodes["012345678901234567890123456789ab"].Name)
	test.Equal(core.NodeTypeMobile, accountGet.State.Nodes["012345678901234567890123456789ab"].NodeType)
	test.Equal("2023-04-15T17:30:40+09:00", accountGet.State.Nodes["012345678901234567890123456789ab"].Timestamp)

	// should be remove invalid record and return nil
	accountName = "get-account-invalid"
	key = test.impl.getKey(accountName)
	accountSet.State.Nodes = nil
	raw, err = json.Marshal(accountSet)
	test.NoError(err)
	err = test.col.KvsSet(key, raw, 0)
	test.NoError(err)
	accountGet, err = test.impl.Get(accountName)
	test.NoError(err)
	test.Nil(accountGet)
	record, err = test.col.KvsGet(key)
	test.NoError(err)
	test.True(record.IsNil())
}

func (test *accountKvsTest) TestSet() {
	// fail when set invalid record
	KEY := "set-account-invalid"
	KEY_RAW := test.impl.getKey(KEY)
	accountSet := &core.Account{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypeAccount,
			Name:              KEY,
			Owner:             KEY,
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              core.GenerateAccountUuid(KEY),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &core.AccountState{
			Pods: nil, // should not be nil
			Nodes: map[string]core.AccountNodeState{
				"012345678901234567890123456789ab": {
					Name:      "test-node",
					NodeType:  core.NodeTypeMobile,
					Timestamp: "2023-04-15T17:30:40+09:00",
					Latitude:  math.NaN(),
					Longitude: math.NaN(),
					Altitude:  math.NaN(),
				},
			},
		},
	}
	err := test.impl.Set(accountSet)
	test.Error(err)
	_, err = test.col.KvsGet(KEY_RAW)
	test.Error(err)

	// can set with the valid record
	accountSet.State.Pods = make(map[string]core.AccountPodState)
	err = test.impl.Set(accountSet)
	defer test.impl.Delete(KEY)
	test.NoError(err)
	accountGet, err := test.impl.Get(KEY)
	test.NoError(err)
	test.Equal(KEY, accountGet.Meta.Name)
}

func (test *accountKvsTest) TestDelete() {
	KEY := "delete-account"
	KEY_RAW := test.impl.getKey(KEY)
	test.impl.Delete(KEY)
	raw, err := test.col.KvsGet(KEY_RAW)
	test.NoError(err)
	test.True(raw.IsNil())
}
