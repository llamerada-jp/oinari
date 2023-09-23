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
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/messaging/driver"
)

type NodeState struct {
	Name      string       `json:"name"`
	ID        string       `json:"id"`
	Account   string       `json:"account"`
	NodeType  api.NodeType `json:"nodeType"`
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Altitude  float64      `json:"altitude"`
}

type NodeRecord struct {
	timestamp time.Time
	state     NodeState
}

type NodeController interface {
	GetNid() string
	GetNodeState() *api.AccountNodeState
	ReceivePublishingNode(state NodeState) error
	SetPosition(latitude, longitude, altitude float64) error
	SetPublicity(r float64) error
	ListNode() ([]NodeState, error)
}

type nodeControllerImpl struct {
	mtx       sync.Mutex
	col       colonio.Colonio
	messaging driver.MessagingDriver
	account   string
	nodeID    string
	nodeName  string
	nodeType  api.NodeType
	nodes     map[string]NodeRecord
	publicity float64
	latitude  float64
	longitude float64
	altitude  float64
}

const (
	NODE_PUBLISH_INTERVAL = 30 * time.Second
	NODE_RECORD_LIFETIME  = 90 * time.Second
)

func NewNodeController(ctx context.Context, col colonio.Colonio, messaging driver.MessagingDriver, account, nodeName string, nodeType api.NodeType) NodeController {
	impl := &nodeControllerImpl{
		col:       col,
		messaging: messaging,
		account:   account,
		nodeID:    col.GetLocalNid(),
		nodeName:  nodeName,
		nodeType:  nodeType,
		nodes:     make(map[string]NodeRecord),
		publicity: 0,
		latitude:  math.NaN(),
		longitude: math.NaN(),
		altitude:  math.NaN(),
	}

	go func() {
		ticker := time.NewTicker(NODE_PUBLISH_INTERVAL)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				impl.cleanup()
				if err := impl.publish(); err != nil {
					log.Printf("publish method on the node controller failed: %s", err.Error())
				}
			}
		}
	}()

	return impl
}

func (impl *nodeControllerImpl) GetNid() string {
	return impl.nodeID
}

func (impl *nodeControllerImpl) GetNodeState() *api.AccountNodeState {
	return &api.AccountNodeState{
		Name:      impl.nodeName,
		NodeType:  impl.nodeType,
		Latitude:  impl.latitude,
		Longitude: impl.longitude,
		Altitude:  impl.altitude,
	}
}

func (impl *nodeControllerImpl) ReceivePublishingNode(state NodeState) error {
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	impl.nodes[state.ID] = NodeRecord{
		timestamp: time.Now(),
		state:     state,
	}

	return nil
}

func (impl *nodeControllerImpl) SetPosition(latitude, longitude, altitude float64) error {
	if latitude < -90.0 || 90 < latitude {
		return fmt.Errorf("latitude should between -90.0 and 90deg")
	}
	if longitude < -180 || 180 < longitude {
		return fmt.Errorf("longitude should between -180.0 and 180deg")
	}
	_, _, err := impl.col.SetPosition(math.Pi*longitude/180.0, math.Pi*latitude/180.0)
	if err != nil {
		return fmt.Errorf("failed to set position of colonio: %w", err)
	}

	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	impl.latitude = latitude
	impl.longitude = longitude
	impl.altitude = altitude

	return nil
}

func (impl *nodeControllerImpl) SetPublicity(r float64) error {
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	impl.publicity = r
	return nil
}

func (impl *nodeControllerImpl) ListNode() ([]NodeState, error) {
	res := make([]NodeState, 0)
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	for _, record := range impl.nodes {
		res = append(res, record.state)
	}

	return res, nil
}

func (impl *nodeControllerImpl) cleanup() {
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	for nodeID, record := range impl.nodes {
		if record.timestamp.Add(NODE_RECORD_LIFETIME).Before(time.Now()) {
			delete(impl.nodes, nodeID)
		}
	}
}

func (impl *nodeControllerImpl) publish() error {
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	// skip if not published
	if impl.publicity == 0 {
		return nil
	}

	// skip if latitude and longitude have not set
	if math.IsNaN(impl.latitude) || math.IsNaN(impl.longitude) {
		return nil
	}

	err := impl.messaging.PublishNode(impl.publicity, impl.nodeID, impl.nodeName, impl.account, impl.nodeType, impl.latitude, impl.longitude, impl.altitude)
	if err != nil {
		return fmt.Errorf("failed to publish node info: %w", err)
	}

	return nil
}

func (ns NodeState) MarshalJSON() ([]byte, error) {
	type Alias NodeState

	var lat *float64
	var lon *float64
	var alt *float64

	if !math.IsNaN(ns.Latitude) {
		lat = &ns.Latitude
	}
	if !math.IsNaN(ns.Longitude) {
		lon = &ns.Longitude
	}
	if !math.IsNaN(ns.Altitude) {
		alt = &ns.Altitude
	}

	return json.Marshal(&struct {
		Alias
		AliasLatitude  *float64 `json:"latitude,omitempty"`
		AliasLongitude *float64 `json:"longitude,omitempty"`
		AliasAltitude  *float64 `json:"altitude,omitempty"`
	}{
		Alias:          (Alias)(ns),
		AliasLatitude:  lat,
		AliasLongitude: lon,
		AliasAltitude:  alt,
	})
}

func (ns *NodeState) UnmarshalJSON(b []byte) error {
	type Alias NodeState

	aux := &struct {
		*Alias
		AliasLatitude  *float64 `json:"latitude,omitempty"`
		AliasLongitude *float64 `json:"longitude,omitempty"`
		AliasAltitude  *float64 `json:"altitude,omitempty"`
	}{
		Alias: (*Alias)(ns),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	if aux.AliasLatitude != nil {
		ns.Latitude = *aux.AliasLatitude
	} else {
		ns.Latitude = math.NaN()
	}
	if aux.AliasLongitude != nil {
		ns.Longitude = *aux.AliasLongitude
	} else {
		ns.Longitude = math.NaN()
	}
	if aux.AliasAltitude != nil {
		ns.Altitude = *aux.AliasAltitude
	} else {
		ns.Altitude = math.NaN()
	}
	return nil
}
