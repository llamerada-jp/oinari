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
package node

import (
	"fmt"
	"strings"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type LocalResource struct {
	key          string
	resourceType api.ResourceType
	recordRaw    []byte
}

type LocalDatastore interface {
	DeleteResource(key string) error
	GetResources() ([]LocalResource, error)
}

type localDatastore struct {
	col colonio.Colonio
}

func NewLocalDatastore(col colonio.Colonio) LocalDatastore {
	return &localDatastore{
		col: col,
	}
}

func (ld *localDatastore) DeleteResource(key string) error {
	return ld.col.KvsSet(key, nil, 0)
}

func (ld *localDatastore) GetResources() ([]LocalResource, error) {
	resources := make([]LocalResource, 0)

	// to avoid dead-lock of ForeachLocalValue, don't call colonio's method in the callback func
	localData := ld.col.KvsGetLocalData()
	defer localData.Free()

	for _, key := range localData.GetKeys() {
		v, err := localData.GetValue(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get local resource value: %w", err)
		}
		// skip nil value. it is workaround because colonio is not provide feature to delete kvs entry.
		if v.IsNil() {
			continue
		}
		raw, err := v.GetBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to get local resource binary: %w", err)
		}
		resourceEntry := strings.Split(key, "/")
		if len(resourceEntry) != 2 {
			return nil, fmt.Errorf("local value key is not supported format:%s", key)
		}
		resources = append(resources, LocalResource{
			key:          key,
			resourceType: api.ResourceType(resourceEntry[0]),
			recordRaw:    raw,
		})
	}

	return resources, nil
}
