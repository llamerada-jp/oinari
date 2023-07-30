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
	"context"
	"log"
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/misc"
)

const KEEP_ALIVE_INTERVAL = 30 * time.Second
const RESOURCE_KEEP_ALIVE_THRESHOLD = KEEP_ALIVE_INTERVAL * 6
const ACCOUNT_DELETION_THRESHOLD = KEEP_ALIVE_INTERVAL * 60

type AccountController interface {
	GetAccountName() string
	GetAccountPodState() (map[string]api.AccountPodState, error)
	Cleanup(account *api.Account) error
	KeepAlivePods(pods []*api.Pod) error
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

func NewAccountController(ctx context.Context, account, localNid string, accountKvs kvs.AccountKvs) AccountController {
	impl := &accountControllerImpl{
		accountName: account,
		localNid:    localNid,
		accountKvs:  accountKvs,
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
				if err := impl.keepAliveNode(); err != nil {
					log.Println(err)
				}
				impl.cleanLogs()
			}
		}
	}()

	return impl
}

func (impl *accountControllerImpl) GetAccountName() string {
	return impl.accountName
}

func (impl *accountControllerImpl) GetAccountPodState() (map[string]api.AccountPodState, error) {
	acc, err := impl.accountKvs.Get(impl.accountName)
	if err != nil {
		return nil, err
	}

	// TODO check not found
	if acc == nil {
		return make(map[string]api.AccountPodState), nil
	}

	return acc.State.Pods, nil
}

func (impl *accountControllerImpl) KeepAlivePods(pods []*api.Pod) error {
	podsByAccount := make(map[string][]*api.Pod)
	for _, pod := range pods {
		podsByAccount[pod.Meta.Owner] = append(podsByAccount[pod.Meta.Owner], pod)
	}

	for accountName, podSlice := range podsByAccount {
		account, err := impl.getOrCreateAccount(accountName)
		if err != nil {
			return err
		}

		for _, pod := range podSlice {
			account.State.Pods[pod.Meta.Uuid] = api.AccountPodState{
				RunningNode: impl.localNid,
				Timestamp:   misc.GetTimestamp(),
			}
		}

		err = impl.accountKvs.Set(account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (impl *accountControllerImpl) Cleanup(account *api.Account) error {
	if err := account.Validate(); err != nil {
		log.Printf("invalid account found %s and it will be deleted: %v", account.Meta.Name, err)
		return impl.accountKvs.Delete(account.Meta.Name)
	}

	log, ok := impl.logs[account.Meta.Name]
	if !ok {
		log = &logEntry{
			lastUpdated: time.Now(),
			pods:        make(map[string]timestampLog),
			nodes:       make(map[string]timestampLog),
		}
		impl.logs[account.Meta.Name] = log
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
		return impl.accountKvs.Set(account)
	}

	if now.After(log.lastUpdated.Add(ACCOUNT_DELETION_THRESHOLD)) {
		return impl.accountKvs.Delete(account.Meta.Name)
	}

	return nil
}

func (impl *accountControllerImpl) keepAliveNode() error {
	account, err := impl.getOrCreateAccount(impl.accountName)
	if err != nil {
		return err
	}

	account.State.Nodes[impl.localNid] = api.AccountNodeState{
		Timestamp: misc.GetTimestamp(),
	}

	return impl.accountKvs.Set(account)
}

func (impl *accountControllerImpl) cleanLogs() {
	now := time.Now()
	for key, log := range impl.logs {
		if now.After(log.lastChecked.Add(RESOURCE_KEEP_ALIVE_THRESHOLD * 2)) {
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
