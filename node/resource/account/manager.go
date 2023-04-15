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
package account

import (
	"context"
	"log"
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/misc"
)

const KEEP_ALIVE_INTERVAL = 30 * time.Second
const RESOURCE_KEEP_ALIVE_THRESHOLD = KEEP_ALIVE_INTERVAL * 6
const ACCOUNT_DELETION_THRESHOLD = KEEP_ALIVE_INTERVAL * 60

type Manager interface {
	GetAccountName() string
	Cleanup(account *api.Account) error
	KeepAlivePods(pods []*api.Pod) error
}

type managerImpl struct {
	accountName string
	localNid    string
	kvs         KvsDriver
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

func NewManager(ctx context.Context, account, localNid string, kvs KvsDriver) Manager {
	mgr := &managerImpl{
		accountName: account,
		localNid:    localNid,
		kvs:         kvs,
		logs:        make(map[string]*logEntry),
	}

	// kick keep alive for each interval
	go func() {
		ticker := time.NewTicker(KEEP_ALIVE_INTERVAL)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if err := mgr.keepAliveNode(); err != nil {
					log.Println(err)
				}
				mgr.cleanLogs()
			}
		}
	}()

	return mgr
}

func (mgr *managerImpl) GetAccountName() string {
	return mgr.accountName
}

func (mgr *managerImpl) KeepAlivePods(pods []*api.Pod) error {
	podsByAccount := make(map[string][]*api.Pod)
	for _, pod := range pods {
		podsByAccount[pod.Meta.Owner] = append(podsByAccount[pod.Meta.Owner], pod)
	}

	for accountName, podSlice := range podsByAccount {
		account, err := mgr.getOrCreateAccount(accountName)
		if err != nil {
			return err
		}

		for _, pod := range podSlice {
			account.State.Pods[pod.Meta.Uuid] = api.AccountPodState{
				RunningNode: mgr.localNid,
				Timestamp:   misc.GetTimestamp(),
			}
		}

		err = mgr.kvs.set(account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *managerImpl) Cleanup(account *api.Account) error {
	if err := account.Validate(); err != nil {
		log.Printf("invalid account found %s and it will be deleted: %v", account.Meta.Name, err)
		return mgr.kvs.delete(account.Meta.Name)
	}

	log, ok := mgr.logs[account.Meta.Name]
	if !ok {
		log = &logEntry{
			lastUpdated: time.Now(),
			pods:        make(map[string]timestampLog),
			nodes:       make(map[string]timestampLog),
		}
		mgr.logs[account.Meta.Name] = log
	}

	updateAccount := false
	now := time.Now()
	log.lastChecked = now

	for key, pod := range account.State.Pods {
		podLog, ok := log.pods[key]
		if !ok || podLog.timestampStr != pod.Timestamp {
			log.pods[key] = timestampLog{
				lastUpdated:  now,
				timestampStr: pod.Timestamp,
			}
			log.lastUpdated = now
			continue
		}
		if now.After(podLog.lastUpdated.Add(RESOURCE_KEEP_ALIVE_THRESHOLD)) {
			delete(account.State.Pods, key)
			updateAccount = true
			delete(log.pods, key)
			log.lastUpdated = now
		}
	}
	for key, podLog := range log.pods {
		if now.After(podLog.lastUpdated.Add(RESOURCE_KEEP_ALIVE_THRESHOLD * 2)) {
			delete(log.pods, key)
			log.lastUpdated = now
		}
	}

	for key, node := range account.State.Nodes {
		nodeLog, ok := log.nodes[key]
		if !ok || nodeLog.timestampStr != node.Timestamp {
			log.nodes[key] = timestampLog{
				lastUpdated:  now,
				timestampStr: node.Timestamp,
			}
			log.lastUpdated = now
			continue
		}
		if now.After(nodeLog.lastUpdated.Add(RESOURCE_KEEP_ALIVE_THRESHOLD)) {
			delete(account.State.Nodes, key)
			updateAccount = true
			delete(log.nodes, key)
			log.lastUpdated = now
		}
	}
	for key, nodeLog := range log.nodes {
		if now.After(nodeLog.lastUpdated.Add(RESOURCE_KEEP_ALIVE_THRESHOLD * 2)) {
			delete(log.nodes, key)
			log.lastUpdated = now
		}
	}

	if updateAccount {
		return mgr.kvs.set(account)
	}

	if now.After(log.lastUpdated.Add(ACCOUNT_DELETION_THRESHOLD)) {
		return mgr.kvs.delete(account.Meta.Name)
	}

	return nil
}

func (mgr *managerImpl) keepAliveNode() error {
	account, err := mgr.getOrCreateAccount(mgr.accountName)
	if err != nil {
		return err
	}

	account.State.Nodes[mgr.localNid] = api.AccountNodeState{
		Timestamp: misc.GetTimestamp(),
	}

	return mgr.kvs.set(account)
}

func (mgr *managerImpl) cleanLogs() {
	now := time.Now()
	for key, log := range mgr.logs {
		if now.After(log.lastChecked.Add(RESOURCE_KEEP_ALIVE_THRESHOLD * 2)) {
			delete(mgr.logs, key)
		}
	}
}

func (mgr *managerImpl) getOrCreateAccount(accountName string) (*api.Account, error) {
	account, err := mgr.kvs.get(accountName)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &api.Account{
			Meta: &api.ObjectMeta{
				Type:        api.ResourceTypeAccount,
				Name:        accountName,
				Owner:       accountName,
				CreatorNode: mgr.localNid,
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
