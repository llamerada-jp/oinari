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
package driver

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/node/messaging"
)

type MessagingDriver interface {
	PublishNode(r float64, nid, name, account string, nodeType core.NodeType, position *core.Vector3) error
	ReconcileContainer(nid, uuid string) error
}

type messagingDriverImpl struct {
	colonio colonio.Colonio
}

func NewMessagingDriver(col colonio.Colonio) MessagingDriver {
	return &messagingDriverImpl{
		colonio: col,
	}
}

func (d *messagingDriverImpl) ReconcileContainer(nid, podUuid string) error {
	raw, err := json.Marshal(messaging.ReconcileContainer{
		PodUuid: podUuid,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal reconcileContainer message: %w", err)
	}

	_, err = d.colonio.MessagingPost(nid, messaging.MessageNameReconcileContainer, raw, 0)
	if err != nil {
		return fmt.Errorf("failed to post reconcileContainer message: %w", err)
	}

	return nil
}

func (d *messagingDriverImpl) PublishNode(r float64, nid, name, account string, nodeType core.NodeType, position *core.Vector3) error {
	raw, err := json.Marshal(messaging.PublishNode{
		Name:     name,
		ID:       nid,
		Account:  account,
		NodeType: nodeType,
		Position: position,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal publishNode message: %w", err)
	}

	err = d.colonio.SpreadPost(math.Pi*position.X/180, math.Pi*position.Y/180, r, messaging.MessageNamePublishNode, raw, 0)
	if err != nil {
		return fmt.Errorf("failed to spread publishNode message: %w", err)
	}

	return nil
}
