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
	api "github.com/llamerada-jp/oinari/api/three"
	controller "github.com/llamerada-jp/oinari/node/controller/three"
	driver "github.com/llamerada-jp/oinari/node/frontend/driver"
	messaging "github.com/llamerada-jp/oinari/node/messaging/three"
)

func InitMessagingHandler(col colonio.Colonio, threeCtrl controller.ObjectController, fd driver.FrontendDriver) error {
	col.MessagingSetHandler(messaging.MessageNameSpreadObject, func(mr *colonio.MessagingRequest, mrw colonio.MessagingResponseWriter) {
		raw, err := mr.Message.GetBinary()
		defer mrw.Write(nil)
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

			obj, err := threeCtrl.Get(msg.UUID)
			if err != nil {
				log.Printf("failed to get object in spreadObject: %s", err.Error())
				return
			}

			if obj != nil {
				if err := fd.PutObjects([]api.Object{*obj}); err != nil {
					log.Printf("failed to put object: %s", err.Error())
					return
				}
			} else {
				if err := fd.DeleteObjects([]string{msg.UUID}); err != nil {
					log.Printf("failed to delete object: %s", err.Error())
					return
				}
			}
		}(raw)
	})

	return nil
}
