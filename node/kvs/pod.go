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
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type PodKvs interface {
	Create(pod *api.Pod) error
	Update(pod *api.Pod) error
	Get(uuid string) (*api.Pod, error)
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

func (impl *podKvsImpl) Create(pod *api.Pod) error {
	if err := pod.Validate(true); err != nil {
		return fmt.Errorf("failed to create pod data: %w", err)
	}

	key := string(api.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to create pod data: %w", err)
	}

	return impl.col.KvsSet(key, raw, colonio.KvsProhibitOverwrite)
}

func (impl *podKvsImpl) Update(pod *api.Pod) error {
	if err := pod.Validate(true); err != nil {
		return fmt.Errorf("failed to update pod data: %w", err)
	}

	key := string(api.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to update pod data: %w", err)
	}

	return impl.col.KvsSet(key, raw, 0)
}

// return pod
func (impl *podKvsImpl) Get(uuid string) (*api.Pod, error) {
	key := string(api.ResourceTypePod) + "/" + uuid
	val, err := impl.col.KvsGet(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod data: %w", err)
	}

	if val.IsNil() {
		return nil, fmt.Errorf("pod is not exists")
	}

	raw, err := val.GetBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to get pod data: %w", err)
	}

	pod := &api.Pod{}
	err = json.Unmarshal(raw, pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// set nil instead of remove record
func (impl *podKvsImpl) Delete(uuid string) error {
	key := string(api.ResourceTypePod) + "/" + uuid
	// TODO check record before set nil for the record
	if err := impl.col.KvsSet(key, nil, 0); err != nil {
		return fmt.Errorf("failed to delete pod data: %w", err)
	}
	return nil
}
