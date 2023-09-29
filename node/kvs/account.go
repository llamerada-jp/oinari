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

	"errors"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api/core"
)

type AccountKvs interface {
	Get(name string) (*core.Account, error)
	Set(account *core.Account) error
	Delete(name string) error
}

type accountKvsImpl struct {
	col colonio.Colonio
}

func NewAccountKvs(col colonio.Colonio) AccountKvs {
	return &accountKvsImpl{
		col: col,
	}
}

func (impl *accountKvsImpl) Get(name string) (*core.Account, error) {
	key := impl.getKey(name)
	val, err := impl.col.KvsGet(key)
	if err != nil {
		if errors.Is(err, colonio.ErrKvsNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if val.IsNil() {
		return nil, nil
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, err
	}

	account := core.Account{}
	err = json.Unmarshal(raw, &account)
	if err != nil {
		return nil, err
	}

	// delete data if it is invalid
	if err := account.Validate(); err != nil {
		impl.Delete(name)
		return nil, nil
	}

	return &account, nil
}

func (impl *accountKvsImpl) Set(account *core.Account) error {
	if err := account.Validate(); err != nil {
		return err
	}

	raw, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = impl.col.KvsSet(impl.getKey(account.Meta.Name), raw, 0)
	if err != nil {
		return err
	}

	return nil
}

func (impl *accountKvsImpl) Delete(name string) error {
	// colonio does not have delete method on KVS, set nil instead of that
	return impl.col.KvsSet(impl.getKey(name), nil, 0)
}

// use sha256 hash as account's uuid
func (impl *accountKvsImpl) getKey(name string) string {
	return string(core.ResourceTypeAccount) + "/" + core.GenerateAccountUuid(name)
}
