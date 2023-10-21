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
package handler

import (
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	controller "github.com/llamerada-jp/oinari/node/controller/three"
	messaging "github.com/llamerada-jp/oinari/node/messaging/three"
)

func InitMessagingHandler(col colonio.Colonio, threeCtrl controller.ObjectController) error {
	col.SpreadSetHandler(messaging.MessageNameSpreadObject, func(sr *colonio.SpreadRequest) {
		raw, err := sr.Message.GetBinary()
		if err != nil {
			log.Printf("failed to read spreadObject message: %s", err.Error())
			return
		}

		go func(raw []byte) {
			var msg messaging.SpreadObject
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Printf("failed to unmarshal spreadObject message: %s", err.Error())
				return
			}

			if err := threeCtrl.ReceiveSpreadEvent(msg.UUID); err != nil {
				log.Printf("failed to receive spreadObject message: %s", err.Error())
				return
			}
		}(raw)
	})

	return nil
}
