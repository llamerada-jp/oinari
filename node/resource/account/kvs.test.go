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
package account

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
	kvs *kvsDriverImpl
}

func NewAccountKvsTest(ctx context.Context, col colonio.Colonio) suite.TestingSuite {
	return &accountKvsTest{
		col: col,
		kvs: &kvsDriverImpl{
			col: col,
		},
	}
}

func (akt *accountKvsTest) TestGet() {
	// return nil if the record is not exist
	KEY := "get-not-exist"
	KEY_RAW := akt.kvs.getKey(KEY)
	_, err := akt.col.KvsGet(KEY_RAW)
	akt.Error(err)
	account, err := akt.kvs.get(KEY)
	akt.NoError(err)
	akt.Nil(account)

	// return nil if the record is nil
	KEY = "get-account-nil"
	KEY_RAW = akt.kvs.getKey(KEY)
	err = akt.col.KvsSet(KEY_RAW, nil, 0)
	akt.NoError(err)
	record, err := akt.col.KvsGet(KEY_RAW)
	akt.NoError(err)
	akt.True(record.IsNil())
	account, err = akt.kvs.get(KEY)
	akt.NoError(err)
	akt.Nil(account)

	// can get valid account record
	KEY = "get-account-valid"
	KEY_RAW = akt.kvs.getKey(KEY)
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
	akt.NoError(err)
	err = akt.col.KvsSet(KEY_RAW, raw, 0)
	akt.NoError(err)
	akt.Equal(api.ResourceTypeAccount, account.Meta.Type)
	akt.Equal("test-account", account.Meta.Name)
	akt.Equal("test-account", account.Meta.Owner)
	akt.Equal("012345678901234567890123456789ab", account.Meta.CreatorNode)
	akt.Equal(api.GenerateAccountUuid("test-account"), account.Meta.Uuid)
	akt.Equal("2023-04-15T17:30:40+09:00", account.Meta.DeletionTimestamp)
	akt.Len(account.State.Pods, 1)
	akt.Equal("012345678901234567890123456789ab", account.State.Pods[podUuid].RunningNode)
	akt.Equal("2023-04-15T17:30:40+09:00", account.State.Pods[podUuid].Timestamp)
	akt.Len(account.State.Nodes, 1)
	akt.Equal("2023-04-15T17:30:40+09:00", account.State.Nodes["012345678901234567890123456789ab"].Timestamp)

	// should be remove invalid record and return nil
	KEY = "get-account-invalid"
	KEY_RAW = akt.kvs.getKey(KEY)
	account.State.Nodes = nil
	raw, err = json.Marshal(account)
	akt.NoError(err)
	err = akt.col.KvsSet(KEY_RAW, raw, 0)
	akt.NoError(err)
	account, err = akt.kvs.get(KEY)
	akt.NoError(err)
	akt.Nil(account)
	record, err = akt.col.KvsGet(KEY_RAW)
	akt.NoError(err)
	akt.True(record.IsNil())
}

func (akt *accountKvsTest) TestSet() {
	// fail when set invalid record
	KEY := "set-account-invalid"
	KEY_RAW := akt.kvs.getKey(KEY)
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
	err := akt.kvs.set(account)
	akt.Error(err)
	_, err = akt.col.KvsGet(KEY_RAW)
	akt.Error(err)

	// can set with the valid record
	account.State.Pods = make(map[string]api.AccountPodState)
	err = akt.kvs.set(account)
	defer akt.kvs.delete(KEY)
	akt.NoError(err)
	account, err = akt.kvs.get(KEY)
	akt.NoError(err)
	akt.Equal(KEY, account.Meta.Name)
}

func (akt *accountKvsTest) TestDelete() {
	KEY := "delete-account"
	KEY_RAW := akt.kvs.getKey(KEY)
	akt.kvs.delete(KEY)
	raw, err := akt.col.KvsGet(KEY_RAW)
	akt.NoError(err)
	akt.True(raw.IsNil())
}
