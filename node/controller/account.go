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
package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/kvs"
)

const ACCOUNT_STATE_RESOURCE_LIFETIME = 180 * time.Second
const ACCOUNT_LIFETIME = 600 * time.Second

type AccountController interface {
	DealLocalResource(raw []byte) (bool, error)

	GetAccountName() string
	GetPodState() (map[string]api.AccountPodState, error)
	GetNodeState() (map[string]api.AccountNodeState, error)
	UpdatePodAndNodeState(account string, pods map[string]api.AccountPodState, nodeID string, nodeState *api.AccountNodeState) error
}

type accountControllerImpl struct {
	accountName string
	localNid    string
	accountKvs  kvs.AccountKvs
	logs        map[string]*logEntry
}

type logEntry struct {
	lastUpdated time.Time
	lastChecked time.Time
	pods        map[string]timestampLog
	nodes       map[string]timestampLog
}

type timestampLog struct {
	lastUpdated  time.Time
	timestampStr string
}

func NewAccountController(account, localNid string, accountKvs kvs.AccountKvs) AccountController {
	return &accountControllerImpl{
		accountName: account,
		localNid:    localNid,
		accountKvs:  accountKvs,
		logs:        make(map[string]*logEntry),
	}
}

func (impl *accountControllerImpl) DealLocalResource(raw []byte) (bool, error) {
	account := &api.Account{}
	if err := json.Unmarshal(raw, account); err != nil {
		return true, fmt.Errorf("failed to unmarshal account record: %w", err)
	}

	if err := account.Validate(); err != nil {
		return true, fmt.Errorf("failed to validate account record: %w", err)
	}

	impl.cleanLogs(account.Meta.Name)
	logE, ok := impl.logs[account.Meta.Name]
	if !ok {
		logE = &logEntry{
			lastUpdated: time.Now(),
			pods:        make(map[string]timestampLog),
			nodes:       make(map[string]timestampLog),
		}
		impl.logs[account.Meta.Name] = logE
	}

	updateAccount := false
	now := time.Now()
	logE.lastChecked = now

	for key, pod := range account.State.Pods {
		podLog, ok := logE.pods[key]
		if !ok || podLog.timestampStr != pod.Timestamp {
			logE.pods[key] = timestampLog{
				lastUpdated:  now,
				timestampStr: pod.Timestamp,
			}
			logE.lastUpdated = now
			continue
		}
		if now.After(podLog.lastUpdated.Add(ACCOUNT_STATE_RESOURCE_LIFETIME)) {
			delete(account.State.Pods, key)
			updateAccount = true
			logE.lastUpdated = now
		}
	}
	for key, podLog := range logE.pods {
		if now.After(podLog.lastUpdated.Add(ACCOUNT_STATE_RESOURCE_LIFETIME * 2)) {
			delete(logE.pods, key)
			logE.lastUpdated = now
		}
	}

	for key, node := range account.State.Nodes {
		nodeLog, ok := logE.nodes[key]
		if !ok || nodeLog.timestampStr != node.Timestamp {
			logE.nodes[key] = timestampLog{
				lastUpdated:  now,
				timestampStr: node.Timestamp,
			}
			logE.lastUpdated = now
			continue
		}
		if now.After(nodeLog.lastUpdated.Add(ACCOUNT_STATE_RESOURCE_LIFETIME)) {
			delete(account.State.Nodes, key)
			updateAccount = true
			logE.lastUpdated = now
		}
	}
	for key, nodeLog := range logE.nodes {
		if now.After(nodeLog.lastUpdated.Add(ACCOUNT_STATE_RESOURCE_LIFETIME * 2)) {
			delete(logE.nodes, key)
			logE.lastUpdated = now
		}
	}

	if updateAccount {
		return false, impl.accountKvs.Set(account)
	}

	if now.After(logE.lastUpdated.Add(ACCOUNT_LIFETIME)) {
		return true, nil
	}

	return false, nil
}

func (impl *accountControllerImpl) GetAccountName() string {
	return impl.accountName
}

func (impl *accountControllerImpl) GetPodState() (map[string]api.AccountPodState, error) {
	acc, err := impl.accountKvs.Get(impl.accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get account record: %s", err)
	}

	// TODO check not found
	if acc == nil {
		return make(map[string]api.AccountPodState), nil
	}

	return acc.State.Pods, nil
}

func (impl *accountControllerImpl) GetNodeState() (map[string]api.AccountNodeState, error) {
	acc, err := impl.accountKvs.Get(impl.accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get account record: %s", err)
	}

	// TODO check not found
	if acc == nil {
		return make(map[string]api.AccountNodeState), nil
	}

	return acc.State.Nodes, nil
}

func (impl *accountControllerImpl) UpdatePodAndNodeState(account string, pods map[string]api.AccountPodState, nodeID string, nodeState *api.AccountNodeState) error {
	record, err := impl.getOrCreateAccount(account)
	if err != nil {
		return err
	}

	for podUUID, podState := range pods {
		record.State.Pods[podUUID] = podState
	}

	if nodeState != nil {
		record.State.Nodes[nodeID] = *nodeState
	}

	if err := impl.accountKvs.Set(record); err != nil {
		return fmt.Errorf("failed to update account record (%s): %w", account, err)
	}

	return nil
}

func (impl *accountControllerImpl) cleanLogs(exclude string) {
	now := time.Now()
	for key, log := range impl.logs {
		if key == exclude {
			continue
		}
		if now.After(log.lastChecked.Add(ACCOUNT_LIFETIME * 2)) {
			delete(impl.logs, key)
		}
	}
}

func (impl *accountControllerImpl) getOrCreateAccount(accountName string) (*api.Account, error) {
	account, err := impl.accountKvs.Get(accountName)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &api.Account{
			Meta: &api.ObjectMeta{
				Type:        api.ResourceTypeAccount,
				Name:        accountName,
				Owner:       accountName,
				CreatorNode: impl.localNid,
				Uuid:        api.GenerateAccountUuid(accountName),
			},
			State: &api.AccountState{
				Pods:  make(map[string]api.AccountPodState),
				Nodes: make(map[string]api.AccountNodeState),
			},
		}
	}

	return account, nil
}
