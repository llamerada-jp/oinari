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
	"github.com/llamerada-jp/oinari/node/misc"
)

type AccountKvs interface {
	Get(name string) (*core.Account, error)
	Set(account *core.Account) error
	Delete(name string) error
}

type accountKvsImpl struct {
	col         colonio.Colonio
	progressing *misc.UniqueSet
}

func NewAccountKvs(col colonio.Colonio) AccountKvs {
	return &accountKvsImpl{
		col:         col,
		progressing: misc.NewUniqueSet(),
	}
}

func (impl *accountKvsImpl) Get(name string) (*core.Account, error) {
	key := impl.getKey(name)
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

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
		// colonio does not have delete method on KVS, set nil instead of that
		impl.col.KvsSet(key, nil, 0)
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

	key := impl.getKey(account.Meta.Name)
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	err = impl.col.KvsSet(key, raw, 0)
	if err != nil {
		return err
	}

	return nil
}

func (impl *accountKvsImpl) Delete(name string) error {
	key := impl.getKey(name)
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	// colonio does not have delete method on KVS, set nil instead of that
	return impl.col.KvsSet(key, nil, 0)
}

// use sha256 hash as account's uuid
func (impl *accountKvsImpl) getKey(name string) string {
	return string(core.ResourceTypeAccount) + "/" + core.GenerateAccountUuid(name)
}
