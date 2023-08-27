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
	"fmt"
	"log"
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/controller"
	"github.com/llamerada-jp/oinari/node/misc"
)

type Manager interface {
	Start(ctx context.Context) error
}

type manager struct {
	localDs       LocalDatastore
	accountCtrl   controller.AccountController
	containerCtrl controller.ContainerController
	nodeCtrl      controller.NodeController
	podCtrl       controller.PodController
}

func NewManager(ld LocalDatastore, accountCtrl controller.AccountController,
	containerCtrl controller.ContainerController, nodeCtrl controller.NodeController,
	podCtrl controller.PodController) Manager {
	return &manager{
		localDs:       ld,
		accountCtrl:   accountCtrl,
		containerCtrl: containerCtrl,
		nodeCtrl:      nodeCtrl,
		podCtrl:       podCtrl,
	}
}

func (mgr *manager) Start(ctx context.Context) error {
	tickerDealLR := time.NewTicker(3 * time.Second)
	defer tickerDealLR.Stop()
	tickerKeepAlive := time.NewTicker(30 * time.Second)
	defer tickerKeepAlive.Stop()

	// first keepalive
	if err := mgr.keepAlive(); err != nil {
		log.Println(err)
	}

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
		willDelete := false
		switch resource.resourceType {
		case api.ResourceTypePod:
			willDelete, err = mgr.podCtrl.DealLocalResource(resource.recordRaw)
		case api.ResourceTypeAccount:
			willDelete, err = mgr.accountCtrl.DealLocalResource(resource.recordRaw)
		}
		if willDelete {
			err := mgr.localDs.DeleteResource(resource.key)
			if err != nil {
				return fmt.Errorf("failed to delete local resource (%s) : %w", resource.key, err)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to deal local resource (%s) : %w", resource.key, err)
		}
	}

	return nil
}

func (mgr *manager) keepAlive() error {
	localAccount := mgr.accountCtrl.GetAccountName()
	localNodeID := mgr.nodeCtrl.GetNid()

	accPodStates := make(map[string]map[string]api.AccountPodState)
	accPodStates[localAccount] = make(map[string]api.AccountPodState)
	for _, info := range mgr.containerCtrl.GetContainerInfos() {
		podStates, ok := accPodStates[info.Owner]
		if !ok {
			podStates = make(map[string]api.AccountPodState)
			accPodStates[info.Owner] = podStates
		}
		podStates[info.PodUUID] = api.AccountPodState{
			RunningNode: localNodeID,
			Timestamp:   misc.GetTimestamp(),
		}
	}

	for account, podStates := range accPodStates {
		if account == localAccount {
			nodeState := mgr.nodeCtrl.GetNodeState()
			nodeState.Timestamp = misc.GetTimestamp()
			if err := mgr.accountCtrl.UpdatePodAndNodeState(account, podStates, localNodeID, nodeState); err != nil {
				return fmt.Errorf("failed to update account state (%s): %w", account, err)
			}
		} else {
			if err := mgr.accountCtrl.UpdatePodAndNodeState(account, podStates, "", nil); err != nil {
				return fmt.Errorf("failed to update account state (%s): %w", account, err)
			}
		}
	}

	return nil
}
