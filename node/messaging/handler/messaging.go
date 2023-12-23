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
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/node/controller"
	"github.com/llamerada-jp/oinari/node/messaging"
)

func InitMessagingHandler(col colonio.Colonio, containerCtrl controller.ContainerController, nodeCtrl controller.NodeController) error {
	// reconcile container
	col.MessagingSetHandler(messaging.MessageNameReconcileContainer, func(mr *colonio.MessagingRequest, mrw colonio.MessagingResponseWriter) {
		raw, err := mr.Message.GetBinary()
		defer mrw.Write(nil)
		if err != nil {
			log.Printf("failed to read reconcileContainer message: %s", err.Error())
			return
		}

		go func(raw []byte) {
			var msg messaging.ReconcileContainer
			err := json.Unmarshal(raw, &msg)
			if err != nil {
				log.Printf("failed to unmarshal reconcileContainer message: %s", err.Error())
				return
			}

			err = containerCtrl.Reconcile(context.Background(), msg.PodUuid)
			if err != nil {
				log.Printf("failed on ContainerController.Reconcile: %s", err.Error())
				return
			}
		}(raw)
	})

	// publish node
	col.SpreadSetHandler(messaging.MessageNamePublishNode, func(sr *colonio.SpreadRequest) {
		raw, err := sr.Message.GetBinary()
		if err != nil {
			log.Printf("failed to read publishNode message: %s", err.Error())
			return
		}

		go func(raw []byte) {
			var msg messaging.PublishNode
			err := json.Unmarshal(raw, &msg)
			if err != nil {
				log.Printf("failed to unmarshal publishNode message: %s", err.Error())
				return
			}

			err = nodeCtrl.ReceivePublishingNode(controller.NodeState{
				Name:     msg.Name,
				ID:       msg.ID,
				Account:  msg.Account,
				NodeType: msg.NodeType,
				Position: msg.Position,
			})
			if err != nil {
				log.Printf("failed on NodeController.ReceivePublishingNode: %s", err.Error())
				return
			}
		}(raw)
	})
	return nil
}
