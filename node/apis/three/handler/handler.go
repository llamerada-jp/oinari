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
	"fmt"

	"github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	coreCtrl "github.com/llamerada-jp/oinari/node/controller"
	threeCtrl "github.com/llamerada-jp/oinari/node/controller/three"
)

func InitHandler(apiMpx crosslink.MultiPlexer, nodeCtrl coreCtrl.NodeController, objCtrl threeCtrl.ObjectController) {
	mpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("three", mpx)

	// CreateObject
	mpx.SetHandler("createObject", crosslink.NewFuncHandler(func(request *three.CreateObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		podUUID := tags[coreCtrl.ContainerLabelPodUUID]

		// use system position if position is not specified
		if request.Spec.Position == nil {
			pos := nodeCtrl.GetPosition()
			request.Spec.Position = &three.Vector3{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		uuid, err := objCtrl.Create(request.Name, podUUID, request.Spec)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("failed to create object: %s", err.Error()))
			return
		}
		writer.ReplySuccess(&three.CreateObjectResponse{
			UUID: uuid,
		})
	}))

	// UpdateObject
	mpx.SetHandler("updateObject", crosslink.NewFuncHandler(func(request *three.UpdateObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		podUUID := tags[coreCtrl.ContainerLabelPodUUID]
		err := objCtrl.Update(request.UUID, podUUID, request.Spec)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("failed to update object: %s", err.Error()))
			return
		}
		writer.ReplySuccess(&three.UpdateObjectResponse{})
	}))

	// GetObject
	mpx.SetHandler("getObject", crosslink.NewFuncHandler(func(request *three.GetObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		podUUID := tags[coreCtrl.ContainerLabelPodUUID]
		object, err := objCtrl.Get(request.UUID, podUUID)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("failed to get object: %s", err.Error()))
			return
		}
		writer.ReplySuccess(&three.GetObjectResponse{
			Object: object,
		})
	}))

	// DeleteObject
	mpx.SetHandler("deleteObject", crosslink.NewFuncHandler(func(request *three.DeleteObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		podUUID := tags[coreCtrl.ContainerLabelPodUUID]
		err := objCtrl.Delete(request.UUID, podUUID)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("failed to delete object: %s", err.Error()))
			return
		}
		writer.ReplySuccess(&three.DeleteObjectResponse{})
	}))
}
