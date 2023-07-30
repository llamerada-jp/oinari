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
	ld      LocalDatastore
	podCtrl controller.PodController
}

func NewManager(ld LocalDatastore, podCtrl controller.PodController) Manager {
	return &manager{
		ld:      ld,
		podCtrl: podCtrl,
	}
}

func (mgr *manager) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			err := mgr.loop(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (mgr *manager) loop(ctx context.Context) error {
	resources, err := mgr.ld.GetResources()
	if err != nil {
		return err
	}

	for _, resource := range resources {
		switch resource.ResourceType {
		case api.ResourceTypePod:
			err = mgr.podCtrl.DealLocalResource(resource.RecordRaw)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
