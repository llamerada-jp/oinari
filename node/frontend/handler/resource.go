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
package command

import (
	"log"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/controller"
)

type createPodRequest struct {
	Name string       `json:"name"`
	Spec *api.PodSpec `json:"spec"`
}

type createPodResponse struct {
	Digest *controller.ApplicationDigest `json:"digest"`
}

type listPodResponse struct {
	Digests []controller.ApplicationDigest `json:"digests"`
}

type migratePodRequest struct {
	Uuid       string `json:"uuid"`
	TargetNode string `json:"targetNode"`
}

type deletePodRequest struct {
	Uuid string `json:"uuid"`
}

func InitResourceHandler(rootMpx crosslink.MultiPlexer, accCtrl controller.AccountController, containerCtrl controller.ContainerController, nodeCtrl controller.NodeController, podCtrl controller.PodController) {
	mpx := crosslink.NewMultiPlexer()
	rootMpx.SetHandler("resource", mpx)

	mpx.SetHandler("createPod", crosslink.NewFuncHandler(
		func(request *createPodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			digest, err := podCtrl.Create(request.Name, accCtrl.GetAccountName(), nodeCtrl.GetNid(), request.Spec)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(createPodResponse{
				Digest: digest,
			})
		}))

	mpx.SetHandler("listPod", crosslink.NewFuncHandler(
		func(request *interface{}, tags map[string]string, writer crosslink.ResponseWriter) {
			res := listPodResponse{
				Digests: make([]controller.ApplicationDigest, 0),
			}

			uuids := make(map[string]any, 0)

			// get pod uuids bound for the account
			podState, err := accCtrl.GetAccountPodState()
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			for uuid := range podState {
				uuids[uuid] = true
			}

			// get pod uuids running on local node
			for _, uuid := range containerCtrl.GetLocalPodUUIDs() {
				uuids[uuid] = true
			}

			// make pod digest
			for uuid := range uuids {
				pod, err := podCtrl.GetPodData(uuid)
				if err != nil {
					log.Printf("error on get pod info: %s", err.Error())
					continue
				}
				res.Digests = append(res.Digests, controller.ApplicationDigest{
					Name:        pod.Meta.Name,
					Uuid:        uuid,
					RunningNode: pod.Status.RunningNode,
					Owner:       pod.Meta.Owner,
					State:       podCtrl.GetContainerStateMessage(pod),
				})
			}

			writer.ReplySuccess(res)
		}))

	mpx.SetHandler("migratePod", crosslink.NewFuncHandler(
		func(param *migratePodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			err := podCtrl.Migrate(param.Uuid, param.TargetNode)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(nil)
		}))

	mpx.SetHandler("deletePod", crosslink.NewFuncHandler(
		func(param *deletePodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			err := podCtrl.Delete(param.Uuid)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(nil)
		}))
}
