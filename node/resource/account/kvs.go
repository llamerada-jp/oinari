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
	"encoding/json"
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type KvsDriver interface {
	get(name string) (*api.Account, error)
	set(account *api.Account) error
}

type kvsDriverImpl struct {
	name string
	col  colonio.Colonio
}

func NewKvsDriver(col colonio.Colonio) KvsDriver {
	return &kvsDriverImpl{
		col: col,
	}
}

func (kvs *kvsDriverImpl) setMeta(name string) error {
	kvs.name = name

	account, err := kvs.getOrCreate(name)
	if err != nil {
		return err
	}

	if account.Meta.Name != name {
		return fmt.Errorf("account uuid collision")
	}

	return nil
}

func (kvs *kvsDriverImpl) bindPod(pod *api.Pod) error {
	account, err := kvs.getOrCreate(kvs.name)
	if err != nil {
		return err
	}

	node, ok := account.State.Pods[pod.Meta.Uuid]
	if ok && node == pod.Status.RunningNode {
		return nil
	}

	account.State.Pods[pod.Meta.Uuid] = pod.Status.RunningNode
	return kvs.set(account)
}

func (kvs *kvsDriverImpl) getOrCreate(name string) (*api.Account, error) {
	key := kvs.getKey(name)
	val, err := kvs.col.KvsGet(key)
	if err == nil {
		raw, err := val.GetBinary()
		if err != nil {
			return nil, err
		}

		var account api.Account
		err = json.Unmarshal(raw, &account)
		if err != nil {
			return nil, err
		}

		return &account, nil
	}
	// TODO: check does account data exist?

	account := &api.Account{
		Meta: &api.ObjectMeta{
			Type:  api.ResourceTypeAccount,
			Name:  name,
			Owner: name,
			Uuid:  kvs.getUUID(name),
		},
		State: &api.AccountState{
			Pods: make(map[string]string),
		},
	}

	err = kvs.set(account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (kvs *kvsDriverImpl) get(name string) (*api.Account, error) {
	key := kvs.getKey(name)
	val, err := kvs.col.KvsGet(key)
	// TODO: return error if err is not `not found error`
	if err != nil {
		return nil, nil
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, err
	}

	var account api.Account
	err = json.Unmarshal(raw, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (kvs *kvsDriverImpl) set(account *api.Account) error {
	if len(account.Meta.Type) == 0 {
		account.Meta.Type = api.ResourceTypeAccount
	}

	if len(account.Meta.Uuid) == 0 {
		account.Meta.Uuid = kvs.getUUID(account.Meta.Name)
	}

	if err := kvs.check(account); err != nil {
		return err
	}

	raw, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = kvs.col.KvsSet(kvs.getKey(account.Meta.Name), raw, 0)
	if err != nil {
		return err
	}

	return nil
}

// use sha256 hash as account's uuid
func (kvs *kvsDriverImpl) getKey(name string) string {
	return string(api.ResourceTypeAccount) + "/" + kvs.getUUID(name)
}
