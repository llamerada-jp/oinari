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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaValidate(t *testing.T) {
	assert := assert.New(t)

	for _, meta := range []ObjectMeta{
		{
			Type:              ResourceTypeAccount,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		{
			Type:              ResourceTypeAccount,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "2023-04-10T00:00:10Z",
		},
	} {
		assert.NoError(meta.Validate(ResourceTypeAccount))
	}

	for title, meta := range map[string]ObjectMeta{
		"empty type": {
			// Type:              ResourceTypeAccount,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"wrong type": {
			Type:              ResourceTypeNode,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"empty name": {
			Type: ResourceTypeAccount,
			// Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"empty owner": {
			Type: ResourceTypeAccount,
			Name: "name",
			// Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"empty node": {
			Type:  ResourceTypeAccount,
			Name:  "name",
			Owner: "owner",
			// CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"invalid node": {
			Type:              ResourceTypeAccount,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "012345678901234567890123456789012",
			Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"empty uuid": {
			Type:        ResourceTypeAccount,
			Name:        "name",
			Owner:       "owner",
			CreatorNode: "01234567890123456789012345678901",
			// Uuid:              "uuid",
			DeletionTimestamp: "",
		},
		"invalid timestamp": {
			Type:              ResourceTypeAccount,
			Name:              "name",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              "uuid",
			DeletionTimestamp: "2023-04-32T00:00:10Z",
		},
	} {
		assert.Error(meta.Validate(ResourceTypeAccount), title)
	}
}
