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
package core

import "fmt"

type ResourceType string

const (
	ResourceTypeAccount = ResourceType("account")
	ResourceTypeNode    = ResourceType("node")
	ResourceTypePod     = ResourceType("pod")
	ResourceTypeRecord  = ResourceType("record")
)

type ObjectMeta struct {
	Type              ResourceType `json:"type"`
	Name              string       `json:"name"`
	Owner             string       `json:"owner"`
	CreatorNode       string       `json:"creatorNode"`
	Uuid              string       `json:"uuid"`
	DeletionTimestamp string       `json:"deletionTimestamp"`
	// TODO: implement parent parameter to cleanup current resource
	// Parent string `json:"parent"`
}

func (meta *ObjectMeta) Validate(t ResourceType) error {
	if meta.Type != t {
		return fmt.Errorf("type field should be %s", t)
	}

	if len(meta.Name) == 0 {
		return fmt.Errorf("name of the resource should be specify")
	}

	if len(meta.Owner) == 0 {
		return fmt.Errorf("owner of the resource should be specify")
	}

	if len(meta.CreatorNode) == 0 {
		return fmt.Errorf("creator node of the resource should be specify")
	}

	if err := ValidateNodeId(meta.CreatorNode); err != nil {
		return fmt.Errorf("invalid creator node was specified (%s): %w",
			meta.CreatorNode, err)
	}

	if len(meta.Uuid) == 0 {
		return fmt.Errorf("uuid of the resource should be specify")
	}

	if len(meta.DeletionTimestamp) != 0 {
		if err := ValidateTimestamp(meta.DeletionTimestamp); err != nil {
			return fmt.Errorf("invalid deletion timestamp was specified (%s): %w",
				meta.DeletionTimestamp, err)
		}
	}

	return nil
}
