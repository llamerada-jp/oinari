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

	"github.com/llamerada-jp/oinari/api"
	app "github.com/llamerada-jp/oinari/api/app/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/apis/core"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/kvs"
)

func InitHandler(apiMpx crosslink.MultiPlexer, manager *core.Manager, c cri.CRI, podKVS kvs.PodKvs, recordKVS kvs.RecordKvs) {
	mpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("core", mpx)

	mpx.SetHandler("ready", crosslink.NewFuncHandler(func(request *app.ReadyRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		containerID, driver, err := getDriver(tags, manager)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`getDriver` failed on `ready` handler: %s", err.Error()))
			return
		}

		containerList, err := c.ListContainers(&cri.ListContainersRequest{
			Filter: &cri.ContainerFilter{
				ID: containerID,
			},
		})
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`ListContainers` failed on `ready` handler: %s", err.Error()))
			return
		}
		if len(containerList.Containers) == 0 {
			writer.ReplyError(fmt.Sprintf("container not found on `ready` handler"))
			return
		}

		sandboxList, err := c.ListPodSandbox(&cri.ListPodSandboxRequest{
			Filter: &cri.PodSandboxFilter{
				ID: containerList.Containers[0].PodSandboxId,
			},
		})
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`ListPodSandbox` failed on `ready` handler: %s", err.Error()))
		}
		if len(sandboxList.Items) == 0 {
			writer.ReplyError("sandbox not found on `ready` handler")
			return
		}

		podUUID := sandboxList.Items[0].Metadata.UID
		pod, err := podKVS.Get(podUUID)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`podKVS.Get` failed on `ready` handler: %s", err.Error()))
			return
		}
		isInitialize := true
		var containerName string
		for idx, status := range pod.Status.ContainerStatuses {
			if status.ContainerID != containerID {
				continue
			}
			containerName = pod.Spec.Containers[idx].Name
			if status.LastState != nil {
				isInitialize = false
			}
			break
		}

		var record *api.Record
		if !isInitialize {
			record, err = recordKVS.Get(podUUID)
			if err != nil {
				writer.ReplyError(fmt.Sprintf("`recordKVS.Get` failed on `ready` handler: %s", err.Error()))
				return
			}
		}

		go func() {
			if record == nil {
				err = driver.Setup(isInitialize, nil)
			} else {
				entry, ok := record.Data.Entries[containerName]
				if !ok {
					err = driver.Setup(isInitialize, nil)
				} else {
					err = driver.Setup(isInitialize, entry.Record)
				}
			}
			if err != nil {
				// TODO: try to restart container
			}
		}()

		writer.ReplySuccess(&app.ReadyResponse{})
	}))

	mpx.SetHandler("output", crosslink.NewFuncHandler(func(request *app.OutputRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		_, _, err := getDriver(tags, manager)
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`getDriver` failed on `output` handler: %s", err.Error()))
			return
		}

		// TODO: broadcast message to neighbors
		_, err = fmt.Println(string(request.Payload))
		if err != nil {
			writer.ReplyError(fmt.Sprintf("`fmt.Println failed on `output` handler: %s", err.Error()))
			return
		}
		writer.ReplySuccess(&app.OutputResponse{
			Length: len(request.Payload),
		})
	}))
}

func getDriver(tags map[string]string, manager *core.Manager) (string, core.CoreDriver, error) {
	containerID, ok := tags["containerID"]
	if !ok {
		return "", nil, fmt.Errorf("containerID should be set when accessing core handler")
	}

	driver := manager.GetDriver(containerID)
	if driver == nil {
		return "", nil, fmt.Errorf("driver not found for %s", containerID)
	}

	return containerID, driver, nil
}
