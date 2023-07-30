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
	"context"
	"encoding/json"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
	"github.com/stretchr/testify/suite"
)

type accountKvsTest struct {
	suite.Suite
	col colonio.Colonio
	acc *accountKvsImpl
}

func NewAccountKvsTest(ctx context.Context, col colonio.Colonio) suite.TestingSuite {
	return &accountKvsTest{
		col: col,
		acc: &accountKvsImpl{
			col: col,
		},
	}
}

func (test *accountKvsTest) TestGet() {
	// return nil if the record is not exist
	KEY := "get-not-exist"
	KEY_RAW := test.acc.getKey(KEY)
	_, err := test.col.KvsGet(KEY_RAW)
	test.Error(err)
	account, err := test.acc.Get(KEY)
	test.NoError(err)
	test.Nil(account)

	// return nil if the record is nil
	KEY = "get-account-nil"
	KEY_RAW = test.acc.getKey(KEY)
	err = test.col.KvsSet(KEY_RAW, nil, 0)
	test.NoError(err)
	record, err := test.col.KvsGet(KEY_RAW)
	test.NoError(err)
	test.True(record.IsNil())
	account, err = test.acc.Get(KEY)
	test.NoError(err)
	test.Nil(account)

	// can get valid account record
	KEY = "get-account-valid"
	KEY_RAW = test.acc.getKey(KEY)
	podUuid := api.GeneratePodUuid()
	account = &api.Account{
		Meta: &api.ObjectMeta{
			Type:              api.ResourceTypeAccount,
			Name:              "test-account",
			Owner:             "test-account",
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              api.GenerateAccountUuid("test-account"),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &api.AccountState{
			Pods: map[string]api.AccountPodState{
				podUuid: {
					RunningNode: "012345678901234567890123456789ab",
					Timestamp:   "2023-04-15T17:30:40+09:00",
				},
			},
			Nodes: map[string]api.AccountNodeState{
				"012345678901234567890123456789ab": {
					Timestamp: "2023-04-15T17:30:40+09:00",
				},
			},
		},
	}
	raw, err := json.Marshal(account)
	test.NoError(err)
	err = test.col.KvsSet(KEY_RAW, raw, 0)
	test.NoError(err)
	test.Equal(api.ResourceTypeAccount, account.Meta.Type)
	test.Equal("test-account", account.Meta.Name)
	test.Equal("test-account", account.Meta.Owner)
	test.Equal("012345678901234567890123456789ab", account.Meta.CreatorNode)
	test.Equal(api.GenerateAccountUuid("test-account"), account.Meta.Uuid)
	test.Equal("2023-04-15T17:30:40+09:00", account.Meta.DeletionTimestamp)
	test.Len(account.State.Pods, 1)
	test.Equal("012345678901234567890123456789ab", account.State.Pods[podUuid].RunningNode)
	test.Equal("2023-04-15T17:30:40+09:00", account.State.Pods[podUuid].Timestamp)
	test.Len(account.State.Nodes, 1)
	test.Equal("2023-04-15T17:30:40+09:00", account.State.Nodes["012345678901234567890123456789ab"].Timestamp)

	// should be remove invalid record and return nil
	KEY = "get-account-invalid"
	KEY_RAW = test.acc.getKey(KEY)
	account.State.Nodes = nil
	raw, err = json.Marshal(account)
	test.NoError(err)
	err = test.col.KvsSet(KEY_RAW, raw, 0)
	test.NoError(err)
	account, err = test.acc.Get(KEY)
	test.NoError(err)
	test.Nil(account)
	record, err = test.col.KvsGet(KEY_RAW)
	test.NoError(err)
	test.True(record.IsNil())
}

func (test *accountKvsTest) TestSet() {
	// fail when set invalid record
	KEY := "set-account-invalid"
	KEY_RAW := test.acc.getKey(KEY)
	account := &api.Account{
		Meta: &api.ObjectMeta{
			Type:              api.ResourceTypeAccount,
			Name:              KEY,
			Owner:             KEY,
			CreatorNode:       "012345678901234567890123456789ab",
			Uuid:              api.GenerateAccountUuid(KEY),
			DeletionTimestamp: "2023-04-15T17:30:40+09:00",
		},
		State: &api.AccountState{
			Pods:  nil, // should not be nil
			Nodes: make(map[string]api.AccountNodeState),
		},
	}
	err := test.acc.Set(account)
	test.Error(err)
	_, err = test.col.KvsGet(KEY_RAW)
	test.Error(err)

	// can set with the valid record
	account.State.Pods = make(map[string]api.AccountPodState)
	err = test.acc.Set(account)
	defer test.acc.Delete(KEY)
	test.NoError(err)
	account, err = test.acc.Get(KEY)
	test.NoError(err)
	test.Equal(KEY, account.Meta.Name)
}

func (test *accountKvsTest) TestDelete() {
	KEY := "delete-account"
	KEY_RAW := test.acc.getKey(KEY)
	test.acc.Delete(KEY)
	raw, err := test.col.KvsGet(KEY_RAW)
	test.NoError(err)
	test.True(raw.IsNil())
}
