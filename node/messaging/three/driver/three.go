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

	"github.com/llamerada-jp/colonio/go/colonio"
	messaging "github.com/llamerada-jp/oinari/node/messaging/three"
)

type ThreeMessagingDriver interface {
	SpreadObject(uuid string, r, latitude, longitude float64) error
}

type threeMessagingDriverImpl struct {
	col colonio.Colonio
}

func NewThreeMessagingDriver(col colonio.Colonio) ThreeMessagingDriver {
	return &threeMessagingDriverImpl{
		col: col,
	}
}

func (impl *threeMessagingDriverImpl) SpreadObject(uuid string, r, latitude, longitude float64) error {
	raw, err := json.Marshal(messaging.SpreadObject{
		UUID: uuid,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	err = impl.col.SpreadPost(latitude, longitude, r, messaging.MessageNameSpreadObject, raw, 0)
	if err != nil {
		return fmt.Errorf("failed to spread %s message: %w", messaging.MessageNameSpreadObject, err)
	}
	return nil
}
