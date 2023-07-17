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

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type KvsDriver interface {
	get(name string) (*api.Account, error)
	set(account *api.Account) error
	delete(name string) error
}

type kvsDriverImpl struct {
	col colonio.Colonio
}

func NewKvsDriver(col colonio.Colonio) KvsDriver {
	return &kvsDriverImpl{
		col: col,
	}
}

func (kvs *kvsDriverImpl) get(name string) (*api.Account, error) {
	key := kvs.getKey(name)
	val, err := kvs.col.KvsGet(key)
	// TODO: return error if err is not `not found error`
	if err != nil {
		return nil, nil
	}

	if val.IsNil() {
		return nil, nil
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, err
	}

	account := api.Account{}
	err = json.Unmarshal(raw, &account)
	if err != nil {
		return nil, err
	}

	// delete data if it is invalid
	if err := account.Validate(); err != nil {
		kvs.delete(name)
		return nil, nil
	}

	return &account, nil
}

func (kvs *kvsDriverImpl) set(account *api.Account) error {
	if err := account.Validate(); err != nil {
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

func (kvs *kvsDriverImpl) delete(name string) error {
	// colonio does not have delete method on KVS, set nil instead of that
	return kvs.col.KvsSet(kvs.getKey(name), nil, 0)
}

// use sha256 hash as account's uuid
func (kvs *kvsDriverImpl) getKey(name string) string {
	return string(api.ResourceTypeAccount) + "/" + api.GenerateAccountUuid(name)
}
