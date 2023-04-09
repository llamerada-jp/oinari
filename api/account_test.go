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
				Nodes: map[string]string{},
			},
		},
	} {
		assert.NoError(account.Validate())
	}

	// invalid patterns
	for title, account := range map[string]*Account{
		"no meta": {
			State: &AccountState{
				Nodes: map[string]string{},
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
				Nodes: map[string]string{},
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
				Nodes: map[string]string{},
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
			Nodes: map[string]string{},
		},
		{
			Nodes: map[string]string{
				"01234567890123456789012345abcdef": "2021-04-09T14:00:40+09:00",
				"01234567890123456789012345abcde0": "2021-04-09T14:00:40Z",
			},
		},
	} {
		assert.NoError(state.validate())
	}

	// invalid patterns
	for title, state := range map[string]*AccountState{
		"nodes is nil": {
			// Nodes: map[string]string{},
		},
		"invalid node id": {
			Nodes: map[string]string{
				"01234567890123456789012345abcdef": "2021-04-09T14:00:40+09:00",
				"01234567890123456789012345abcdez": "2021-04-09T14:00:40Z",
			},
		},
		"invalid timestamp": {
			Nodes: map[string]string{
				"01234567890123456789012345abcdef": "2021-04-09T14:00:40+09:00",
				"01234567890123456789012345abcde0": "",
			},
		},
	} {
		assert.Error(state.validate(), title)
	}
}
