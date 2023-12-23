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

type RecordKvs interface {
	Get(podUuid string) (*core.Record, error)
	Set(record *core.Record) error
	Delete(podUuid string) error
}

type recordKVSImpl struct {
	col         colonio.Colonio
	progressing *misc.UniqueSet
}

func NewRecordKvs(col colonio.Colonio) RecordKvs {
	return &recordKVSImpl{
		col:         col,
		progressing: misc.NewUniqueSet(),
	}
}

func (impl *recordKVSImpl) Get(podUuid string) (*core.Record, error) {
	key := string(core.ResourceTypeRecord) + "/" + podUuid
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

	record := &core.Record{}
	err = json.Unmarshal(raw, record)
	if err != nil {
		return nil, err
	}

	if err := record.Validate(); err != nil {
		// colonio does not have delete method on KVS, set nil instead of that
		impl.col.KvsSet(key, nil, 0)
		return nil, nil
	}

	return record, nil
}

func (impl *recordKVSImpl) Set(record *core.Record) error {
	if err := record.Validate(); err != nil {
		return err
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return err
	}

	key := string(core.ResourceTypeRecord) + "/" + record.Meta.Uuid
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	err = impl.col.KvsSet(key, raw, 0)
	return err
}

func (impl *recordKVSImpl) Delete(podUuid string) error {
	key := string(core.ResourceTypeRecord) + "/" + podUuid
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	// colonio does not have delete method on KVS, set nil instead of that
	return impl.col.KvsSet(key, nil, 0)
}
