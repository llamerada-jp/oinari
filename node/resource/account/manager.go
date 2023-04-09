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
)

const KEEP_ALIVE_INTERVAL = 30 * time.Second

type Manager interface {
	GetAccountName() string
	BindPod(pod *api.Pod) error
	CheckByLocalData(account *api.Account) error
}

type managerImpl struct {
	accountName string
	localNid    string
	kvs         KvsDriver
}

func NewManager(ctx context.Context, account, localNid string, kvs KvsDriver) Manager {
	mgr := &managerImpl{
		accountName: account,
		localNid:    localNid,
		kvs:         kvs,
	}

	// kick keep alive for each interval
	go func() {
		ticker := time.NewTicker(KEEP_ALIVE_INTERVAL)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if err := mgr.keepAlive(); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return mgr
}

func (mgr *managerImpl) GetAccountName() string {
	return mgr.accountName
}

func (mgr *managerImpl) BindPod(pod *api.Pod) error {
	return mgr.kvs.bindPod(pod)
}

func (mgr *managerImpl) CheckByLocalData(account *api.Account) error {
	log.Fatal("todo")
	return nil
}

func (mgr *managerImpl) keepAlive() error {
	account, err := mgr.kvs.get(mgr.accountName)
	if err != nil {
		return err
	}

	if account == nil {
		account = &api.Account{
			Meta: &api.ObjectMeta{
				Name:  mgr.accountName,
				Owner: mgr.accountName,
			},
			State: &api.AccountState{},
		}
	}

}
