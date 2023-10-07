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
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api/core"
)

type PodKvs interface {
	Create(pod *core.Pod) error
	Update(pod *core.Pod) error
	Get(uuid string) (*core.Pod, error)
	Delete(uuid string) error
}

type podKvsImpl struct {
	col colonio.Colonio
}

func NewPodKvs(col colonio.Colonio) PodKvs {
	return &podKvsImpl{
		col: col,
	}
}

func (impl *podKvsImpl) Create(pod *core.Pod) error {
	if err := pod.Validate(true); err != nil {
		return fmt.Errorf("failed to create pod data: %w", err)
	}

	key := string(core.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to create pod data: %w", err)
	}

	val, err := impl.col.KvsGet(key)
	if errors.Is(err, colonio.ErrKvsNotFound) {
		err = impl.col.KvsSet(key, raw, colonio.KvsProhibitOverwrite)
		if err != nil {
			return fmt.Errorf("failed to set pod data with prohibit overwrite option: %w", err)
		}
		return nil
	}

	// TODO: this implement might occur collision, fix this after delete feature is implemented in colonio
	if val.IsNil() {
		err = impl.col.KvsSet(key, raw, 0)
		if err != nil {
			return fmt.Errorf("failed to set pod data: %w", err)
		}
		return nil
	}

	return fmt.Errorf("there is an duplicate uuid pod data")
}

func (impl *podKvsImpl) Update(pod *core.Pod) error {
	if err := pod.Validate(true); err != nil {
		return fmt.Errorf("failed to update pod data: %w", err)
	}

	key := string(core.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to update pod data: %w", err)
	}

	val, err := impl.col.KvsGet(key)
	if err != nil {
		return fmt.Errorf("failed to check for the existence of the data: %w", err)
	}
	if val.IsNil() {
		return fmt.Errorf("the data might be deleted")
	}

	return impl.col.KvsSet(key, raw, 0)
}

// return pod
func (impl *podKvsImpl) Get(uuid string) (*core.Pod, error) {
	key := string(core.ResourceTypePod) + "/" + uuid
	val, err := impl.col.KvsGet(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data: %w", err)
	}

	if val.IsNil() {
		return nil, fmt.Errorf("the record is not exists")
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, fmt.Errorf("invalid raw data format: %w", err)
	}

	pod := &core.Pod{}
	err = json.Unmarshal(raw, pod)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	return pod, nil
}

// set nil instead of remove record
func (impl *podKvsImpl) Delete(uuid string) error {
	key := string(core.ResourceTypePod) + "/" + uuid
	// TODO check record before set nil for the record
	if err := impl.col.KvsSet(key, nil, 0); err != nil {
		return fmt.Errorf("failed to delete the record: %w", err)
	}
	return nil
}
