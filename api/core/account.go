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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/exp/slices"
)

type Account struct {
	Meta  *ObjectMeta   `json:"meta"`
	State *AccountState `json:"state"`
}

type AccountState struct {
	// map describing pod's uuid and pod state
	Pods map[string]AccountPodState `json:"pods"`
	// map describing nid and timestamp of keepalive
	Nodes map[string]AccountNodeState `json:"nodes"`
}

type AccountPodState struct {
	RunningNode string `json:"runningNode"`
	Timestamp   string `json:"timestamp"`
}

type NodeType string

const (
	NodeTypeMobile      NodeType = "Mobile"
	NodeTypeSmallDevice NodeType = "SmallDevice"
	NodeTypePC          NodeType = "PC"
	NodeTypeServer      NodeType = "Server"
	NodeTypeGrass       NodeType = "Grass"
	NodeTypeOther       NodeType = "Other"
)

var NodeTypeAccepted = []NodeType{
	NodeTypeMobile,
	NodeTypeSmallDevice,
	NodeTypePC,
	NodeTypeServer,
	NodeTypeGrass,
	NodeTypeOther,
}

type AccountNodeState struct {
	Name      string   `json:"name"`
	Timestamp string   `json:"timestamp"`
	NodeType  NodeType `json:"nodeType"`
	Position  *Vector3 `json:"position,omitempty"`
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

	if err := account.Meta.Validate(ResourceTypeAccount); err != nil {
		return fmt.Errorf("invalid metadata for %s %w", account.Meta.Name, err)
	}

	if account.Meta.Uuid != GenerateAccountUuid(account.Meta.Name) {
		return fmt.Errorf("invalid uuid for %s", account.Meta.Name)
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
	if state.Pods == nil {
		return fmt.Errorf("pods field should be filled")
	}

	for podUuid, podState := range state.Pods {
		if err := ValidatePodUuid(podUuid); err != nil {
			return fmt.Errorf("there is an invalid pod uuid for pods: %w", err)
		}
		if len(podState.RunningNode) != 0 {
			if err := ValidateNodeId(podState.RunningNode); err != nil {
				return fmt.Errorf("there is an invalid node if for pod %s: %w", podUuid, err)
			}
		}
		if err := ValidateTimestamp(podState.Timestamp); err != nil {
			return fmt.Errorf("there is an invalid timestamp for pod %s: %w", podUuid, err)
		}
	}

	if state.Nodes == nil {
		return fmt.Errorf("nodes field should be filled")
	}

	for nid, nodeState := range state.Nodes {
		if err := ValidateNodeId(nid); err != nil {
			return fmt.Errorf("there is an invalid node id for nodes: %w", err)
		}
		if err := ValidateTimestamp(nodeState.Timestamp); err != nil {
			return fmt.Errorf("there is an invalid timestamp for nodes: %w", err)
		}
		if !slices.Contains(NodeTypeAccepted, nodeState.NodeType) {
			return fmt.Errorf("there is an unsupported node type in the node state")
		}
	}

	return nil
}
