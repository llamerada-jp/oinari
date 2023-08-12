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
package node

import (
	"context"
	"log"
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/controller"
)

type Manager interface {
	Start(ctx context.Context) error
}

type manager struct {
	localDs     LocalDatastore
	accountCtrl controller.AccountController
	podCtrl     controller.PodController
}

func NewManager(ld LocalDatastore, accountCtrl controller.AccountController, podCtrl controller.PodController) Manager {
	return &manager{
		localDs:     ld,
		accountCtrl: accountCtrl,
		podCtrl:     podCtrl,
	}
}

func (mgr *manager) Start(ctx context.Context) error {
	tickerDealLR := time.NewTicker(3 * time.Second)
	defer tickerDealLR.Stop()
	tickerKeepAlive := time.NewTicker(30 * time.Second)
	defer tickerKeepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-tickerDealLR.C:
			if err := mgr.dealLocalResource(); err != nil {
				log.Println(err)
			}

		case <-tickerKeepAlive.C:
			if err := mgr.keepAlive(); err != nil {
				log.Println(err)
			}

		}
	}
}

func (mgr *manager) dealLocalResource() error {
	resources, err := mgr.localDs.GetResources()
	if err != nil {
		return err
	}

	for _, resource := range resources {
		switch resource.ResourceType {
		case api.ResourceTypePod:
			err = mgr.podCtrl.DealLocalResource(resource.RecordRaw)
		case api.ResourceTypeAccount:
			err = mgr.accountCtrl.DealLocalResource(resource.RecordRaw)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *manager) keepAlive() error {
}
