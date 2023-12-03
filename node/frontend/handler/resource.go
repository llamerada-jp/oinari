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

	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/controller"
	"github.com/llamerada-jp/oinari/node/misc"
)

type setPositionRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type SetPublicityRequest struct {
	Range float64 `json:"range"`
}

type ListNodeResponse struct {
	Nodes []controller.NodeState `json:"nodes"`
}

type createPodRequest struct {
	Name string        `json:"name"`
	Spec *core.PodSpec `json:"spec"`
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

type configRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func InitResourceHandler(nodeMpx crosslink.MultiPlexer, accCtrl controller.AccountController, containerCtrl controller.ContainerController, nodeCtrl controller.NodeController, podCtrl controller.PodController) {
	mpx := crosslink.NewMultiPlexer()
	nodeMpx.SetHandler("resource", mpx)

	// node resource
	mpx.SetHandler("setNodePosition", crosslink.NewFuncHandler(func(request *setPositionRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := nodeCtrl.SetPosition(request.Latitude, request.Longitude, request.Altitude)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("setNodePublicity", crosslink.NewFuncHandler(func(request *SetPublicityRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := nodeCtrl.SetPublicity(request.Range)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("listNode", crosslink.NewFuncHandler(func(request *interface{}, tags map[string]string, writer crosslink.ResponseWriter) {
		nodes, err := nodeCtrl.ListNode()
		if err != nil {
			writer.ReplyError(err.Error())
		}

		ids := make(map[string]bool)
		for _, node := range nodes {
			ids[node.ID] = true
		}

		nodeState, err := accCtrl.GetNodeState()
		if err != nil {
			writer.ReplyError(err.Error())
		}
		account := accCtrl.GetAccountName()
		for nodeID, state := range nodeState {
			// skip duplicate node
			if ids[nodeID] == true {
				continue
			}
			nodes = append(nodes, controller.NodeState{
				Name:      state.Name,
				ID:        nodeID,
				Account:   account,
				NodeType:  state.NodeType,
				Latitude:  state.Latitude,
				Longitude: state.Longitude,
				Altitude:  state.Altitude,
			})
		}

		res := ListNodeResponse{
			Nodes: nodes,
		}

		writer.ReplySuccess(res)
	}))

	// pod resource
	mpx.SetHandler("createPod", crosslink.NewFuncHandler(
		func(request *createPodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			owner := accCtrl.GetAccountName()
			digest, err := podCtrl.Create(request.Name, owner, nodeCtrl.GetNid(), request.Spec)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}

			err = accCtrl.UpdatePodAndNodeState(owner, map[string]core.AccountPodState{
				digest.Uuid: {
					RunningNode: "",
					Timestamp:   misc.GetTimestamp(),
				},
			}, "", nil)
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
			podState, err := accCtrl.GetPodState()
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			for uuid := range podState {
				uuids[uuid] = true
			}

			// get pod uuids running on local node
			for _, info := range containerCtrl.GetContainerInfos() {
				uuids[info.PodUUID] = true
			}

			// make pod digest
			for uuid := range uuids {
				pod, err := podCtrl.GetPodData(uuid)
				if err != nil {
					log.Printf("error on get pod info: %s", err.Error())
					continue
				}
				res.Digests = append(res.Digests, controller.ApplicationDigest{
					Name:          pod.Meta.Name,
					Uuid:          uuid,
					RunningNodeID: pod.Status.RunningNode,
					Owner:         pod.Meta.Owner,
					State:         podCtrl.GetContainerStateMessage(pod),
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
