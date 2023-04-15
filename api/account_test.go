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
package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAccountUuid(t *testing.T) {
	assert := assert.New(t)

	id1 := GenerateAccountUuid("name1")
	assert.NotEmpty(id1)
	assert.Equal(GenerateAccountUuid("name1"), id1)

	id2 := GenerateAccountUuid("name2")
	assert.NotEmpty(id2)
	assert.NotEqual(id1, id2)
}

func TestAccountValidate(t *testing.T) {
	assert := assert.New(t)

	// valid patterns
	for _, account := range []*Account{
		{
			Meta: &ObjectMeta{
				Type:              ResourceTypeAccount,
				Name:              "name",
				Owner:             "owner",
				CreatorNode:       "01234567890123456789012345abcdef",
				Uuid:              GenerateAccountUuid("name"),
				DeletionTimestamp: "",
			},
			State: &AccountState{
				Pods:  make(map[string]AccountPodState),
				Nodes: make(map[string]AccountNodeState),
			},
		},
	} {
		assert.NoError(account.Validate())
	}

	// invalid patterns
	for title, account := range map[string]*Account{
		"no meta": {
			State: &AccountState{
				Pods:  make(map[string]AccountPodState),
				Nodes: make(map[string]AccountNodeState),
			},
		},
		"invalid meta": {
			Meta: &ObjectMeta{
				Type:              ResourceTypeNode,
				Name:              "name",
				Owner:             "owner",
				CreatorNode:       "01234567890123456789012345abcdef",
				Uuid:              GenerateAccountUuid("name"),
				DeletionTimestamp: "",
			},
			State: &AccountState{
				Pods:  make(map[string]AccountPodState),
				Nodes: make(map[string]AccountNodeState),
			},
		},
		"invalid uuid": {
			Meta: &ObjectMeta{
				Type:              ResourceTypeAccount,
				Name:              "name",
				Owner:             "owner",
				CreatorNode:       "01234567890123456789012345abcdef",
				Uuid:              GenerateAccountUuid("wrong name"),
				DeletionTimestamp: "",
			},
			State: &AccountState{
				Pods:  make(map[string]AccountPodState),
				Nodes: make(map[string]AccountNodeState),
			},
		},
		"no state": {
			Meta: &ObjectMeta{
				Type:              ResourceTypeAccount,
				Name:              "name",
				Owner:             "owner",
				CreatorNode:       "01234567890123456789012345abcdef",
				Uuid:              GenerateAccountUuid("name"),
				DeletionTimestamp: "",
			},
		},
		"invalid state": {
			Meta: &ObjectMeta{
				Type:              ResourceTypeAccount,
				Name:              "name",
				Owner:             "owner",
				CreatorNode:       "01234567890123456789012345abcdef",
				Uuid:              GenerateAccountUuid("name"),
				DeletionTimestamp: "",
			},
			State: &AccountState{},
		},
	} {
		assert.Error(account.Validate(), title)
	}
}

func TestAccountStateValidate(t *testing.T) {
	assert := assert.New(t)

	// valid patterns
	for _, state := range []*AccountState{
		{
			Pods:  make(map[string]AccountPodState),
			Nodes: make(map[string]AccountNodeState),
		},
		{
			Pods: map[string]AccountPodState{
				GeneratePodUuid(): {
					RunningNode: "01234567890123456789012345abcdef",
					Timestamp:   "2021-04-09T14:00:40+09:00",
				},
			},
			Nodes: map[string]AccountNodeState{
				"01234567890123456789012345abcdef": {
					Timestamp: "2021-04-09T14:00:40+09:00",
				},
				"01234567890123456789012345abcde0": {
					Timestamp: "2021-04-09T14:00:40Z",
				},
			},
		},
	} {
		assert.NoError(state.validate())
	}

	// invalid patterns
	for title, state := range map[string]*AccountState{
		"pods is nil": {
			// Pods: make(map[string]AccountPodState),
			Nodes: make(map[string]AccountNodeState),
		},
		"invalid pod uuid": {
			Pods: map[string]AccountPodState{
				GeneratePodUuid() + "-": {
					RunningNode: "01234567890123456789012345abcdef",
					Timestamp:   "2021-04-09T14:00:40+09:00",
				},
			},
			Nodes: make(map[string]AccountNodeState),
		},
		"invalid node id for the pod": {
			Pods: map[string]AccountPodState{
				GeneratePodUuid(): {
					RunningNode: "01234567890123456789012345abcdef+",
					Timestamp:   "2021-04-09T14:00:40+09:00",
				},
			},
			Nodes: make(map[string]AccountNodeState),
		},
		"invalid timestamp for the pod": {
			Pods: map[string]AccountPodState{
				GeneratePodUuid(): {
					RunningNode: "01234567890123456789012345abcdef+",
					Timestamp:   "2021-04-09T14:00:40+09:00U",
				},
			},
			Nodes: make(map[string]AccountNodeState),
		},

		"nodes is nil": {
			Pods: make(map[string]AccountPodState),
			// Nodes: make(map[string]AccountNodeState),
		},
		"invalid node id": {
			Pods: make(map[string]AccountPodState),
			Nodes: map[string]AccountNodeState{
				"01234567890123456789012345abcdef": {
					Timestamp: "2021-04-09T14:00:40+09:00",
				},
				"01234567890123456789012345abcdez": {
					Timestamp: "2021-04-09T14:00:40Z",
				},
			},
		},
		"invalid timestamp for the node": {
			Pods: make(map[string]AccountPodState),
			Nodes: map[string]AccountNodeState{
				"01234567890123456789012345abcdef": {
					Timestamp: "2021-04-09T14:00:40+09:00",
				},
				"01234567890123456789012345abcde0": {
					Timestamp: "",
				},
			},
		},
	} {
		assert.Error(state.validate(), title)
	}
}
