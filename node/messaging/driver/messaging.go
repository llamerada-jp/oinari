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

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/node/messaging"
)

type MessagingDriver interface {
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
	raw, err := json.Marshal(messaging.VitalizePod{
		PodUuid: podUuid,
	})
	if err != nil {
		return err
	}
	_, err = d.colonio.MessagingPost(nid, "reconcileContainer", raw, 0)
	if err != nil {
		return err
	}
	return nil
}
