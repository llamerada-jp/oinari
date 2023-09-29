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
package mock

import (
	"sync"

	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/node/messaging"
	"github.com/llamerada-jp/oinari/node/messaging/driver"
)

type MessagingRecord struct {
	DestNodeID         string
	DestR              float64
	DestLat            float64
	DestLon            float64
	PublishNode        *messaging.PublishNode
	ReconcileContainer *messaging.ReconcileContainer
}

type MessagingDriver struct {
	mutex   sync.Mutex
	Records []*MessagingRecord
}

var _ driver.MessagingDriver = &MessagingDriver{}

func NewMessagingDriverMock() *MessagingDriver {
	return &MessagingDriver{
		Records: make([]*MessagingRecord, 0),
	}
}

func (md *MessagingDriver) ResetRecord() {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	md.Records = make([]*MessagingRecord, 0)
}

func (md *MessagingDriver) PublishNode(r float64, nid, name, account string, nodeType core.NodeType, latitude, longitude, altitude float64) error {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	md.Records = append(md.Records, &MessagingRecord{
		DestR:   r,
		DestLat: latitude,
		DestLon: longitude,
		PublishNode: &messaging.PublishNode{
			Name:      name,
			ID:        nid,
			Account:   account,
			NodeType:  nodeType,
			Latitude:  latitude,
			Longitude: longitude,
			Altitude:  altitude,
		},
	})

	return nil
}

func (md *MessagingDriver) ReconcileContainer(nid, podUuid string) error {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	md.Records = append(md.Records, &MessagingRecord{
		DestNodeID: nid,
		ReconcileContainer: &messaging.ReconcileContainer{
			PodUuid: podUuid,
		},
	})

	return nil
}
