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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Account struct {
	Meta  *ObjectMeta   `json:"meta"`
	State *AccountState `json:"state"`
}

type AccountState struct {
	//Pods map[string]string `json:"pods"`
	// map describing nid and timestamp of keepalive
	Nodes map[string]string `json:"nodes"`
}

// use sha256 hash as account's uuid
func GenerateAccountUuid(name string) string {
	hash := sha256.Sum256([]byte(name))
	return hex.EncodeToString(hash[:])
}

func (account *Account) Validate() error {
	if account.Meta == nil {
		return fmt.Errorf("metadata field should be filled")
	}

	if err := account.Meta.Validate(ResourceTypeAccount, GenerateAccountUuid(account.Meta.Name)); err != nil {
		return fmt.Errorf("invalid metadata for %s %w", account.Meta.Name, err)
	}

	if account.State == nil {
		return fmt.Errorf("state filed should be filled")
	}

	if err := account.State.validate(); err != nil {
		return fmt.Errorf("invalid account state for %s %w", account.Meta.Name, err)
	}
	return nil
}

func (state *AccountState) validate() error {
	if state.Nodes == nil {
		return fmt.Errorf("nodes field should be filled")
	}

	for nid, timestamp := range state.Nodes {
		if err := ValidateNodeId(nid); err != nil {
			return fmt.Errorf("there is an invalid node id for nodes: %w", err)
		}
		if err := ValidateTimestamp(timestamp); err != nil {
			return fmt.Errorf("there is an invalid timestamp for nodes: %w", err)
		}
	}

	return nil
}
