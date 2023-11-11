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
package three

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	api "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/node/misc"
)

type ObjectKVS interface {
	Create(object *api.Object) error
	Update(object *api.Object) error
	Get(uuid string) (*api.Object, error)
	Delete(uuid string) error
}

type objectKVSImpl struct {
	col         colonio.Colonio
	progressing *misc.UniqueSet
}

func NewObjectKVS(col colonio.Colonio) ObjectKVS {
	return &objectKVSImpl{
		col:         col,
		progressing: misc.NewUniqueSet(),
	}
}

func (impl *objectKVSImpl) Create(object *api.Object) error {
	if err := object.Validate(); err != nil {
		return fmt.Errorf("object record invalid: %w", err)
	}

	key := string(api.ResourceTypeThreeObject) + "/" + object.Meta.Uuid
	raw, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	val, err := impl.col.KvsGet(key)
	if errors.Is(err, colonio.ErrKvsNotFound) {
		err = impl.col.KvsSet(key, raw, colonio.KvsProhibitOverwrite)
		if err != nil {
			return fmt.Errorf("failed to get existing data: %w", err)
		}
		return nil
	}

	// TODO: this implement might occur collision, fix this after delete feature is implemented in colonio
	if val.IsNil() {
		err = impl.col.KvsSet(key, raw, 0)
		if err != nil {
			return fmt.Errorf("failed to set raw data: %w", err)
		}
		return nil
	}

	return fmt.Errorf("there is an duplicate uuid object data")
}

func (impl *objectKVSImpl) Update(object *api.Object) error {
	if err := object.Validate(); err != nil {
		return fmt.Errorf("object record invalid: %w", err)
	}

	key := string(api.ResourceTypeThreeObject) + "/" + object.Meta.Uuid
	raw, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	val, err := impl.col.KvsGet(key)
	if err != nil {
		return fmt.Errorf("failed to check for the existence of the data: %w", err)
	}
	if val.IsNil() {
		return fmt.Errorf("the data might be deleted")
	}

	return impl.col.KvsSet(key, raw, 0)
}

func (impl *objectKVSImpl) Get(uuid string) (*api.Object, error) {
	key := string(api.ResourceTypeThreeObject) + "/" + uuid
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	val, err := impl.col.KvsGet(key)
	if err != nil {
		if errors.Is(err, colonio.ErrKvsNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get raw data: %w", err)
	}

	if val.IsNil() {
		return nil, nil
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, fmt.Errorf("invalid raw data format: %w", err)
	}

	object := &api.Object{}
	err = json.Unmarshal(raw, object)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	return object, nil
}

func (impl *objectKVSImpl) Delete(uuid string) error {
	key := string(api.ResourceTypeThreeObject) + "/" + uuid
	impl.progressing.Insert(key)
	defer impl.progressing.Remove(key)

	// TODO check record before set nil for the record
	if err := impl.col.KvsSet(key, nil, 0); err != nil {
		return fmt.Errorf("failed to delete the record: %w", err)
	}
	return nil
}
