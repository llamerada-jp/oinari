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
package cri

import (
	"encoding/json"
	"strings"

	"github.com/llamerada-jp/oinari/lib/crosslink"
)

const (
	crosslinkPath = "cri"
)

type criImpl struct {
	cl crosslink.Crosslink
}

func NewCRI(cl crosslink.Crosslink) CRI {
	return &criImpl{
		cl: cl,
	}
}

func criCallHelper[REQ any, RES any](ci *criImpl, path string, request *REQ) (*RES, error) {
	ch := make(chan *RES)
	var funcError error

	ci.cl.Call(strings.Join([]string{crosslinkPath, path}, "/"), request, nil,
		func(response []byte, err error) {
			defer close(ch)

			if err != nil {
				funcError = err
				return
			}

			var res RES
			err = json.Unmarshal(response, &res)
			if err != nil {
				funcError = err
				return
			}

			ch <- &res
		})

	res, ok := <-ch
	if !ok {
		return nil, funcError
	}
	return res, nil
}

func (ci *criImpl) RunPodSandbox(request *RunPodSandboxRequest) (*RunPodSandboxResponse, error) {
	return criCallHelper[RunPodSandboxRequest, RunPodSandboxResponse](ci, "runPodSandbox", request)
}

func (ci *criImpl) StopPodSandbox(request *StopPodSandboxRequest) (*StopPodSandboxResponse, error) {
	return criCallHelper[StopPodSandboxRequest, StopPodSandboxResponse](ci, "stopPodSandbox", request)
}

func (ci *criImpl) RemovePodSandbox(request *RemovePodSandboxRequest) (*RemovePodSandboxResponse, error) {
	return criCallHelper[RemovePodSandboxRequest, RemovePodSandboxResponse](ci, "removePodSandbox", request)
}

func (ci *criImpl) PodSandboxStatus(request *PodSandboxStatusRequest) (*PodSandboxStatusResponse, error) {
	return criCallHelper[PodSandboxStatusRequest, PodSandboxStatusResponse](ci, "podSandboxStatus", request)
}

func (ci *criImpl) ListPodSandbox(request *ListPodSandboxRequest) (*ListPodSandboxResponse, error) {
	return criCallHelper[ListPodSandboxRequest, ListPodSandboxResponse](ci, "listPodSandbox", request)
}

func (ci *criImpl) CreateContainer(request *CreateContainerRequest) (*CreateContainerResponse, error) {
	return criCallHelper[CreateContainerRequest, CreateContainerResponse](ci, "createContainer", request)
}

func (ci *criImpl) StartContainer(request *StartContainerRequest) (*StartContainerResponse, error) {
	return criCallHelper[StartContainerRequest, StartContainerResponse](ci, "startContainer", request)
}

func (ci *criImpl) StopContainer(request *StopContainerRequest) (*StopContainerResponse, error) {
	return criCallHelper[StopContainerRequest, StopContainerResponse](ci, "stopContainer", request)
}

func (ci *criImpl) RemoveContainer(request *RemoveContainerRequest) (*RemoveContainerResponse, error) {
	return criCallHelper[RemoveContainerRequest, RemoveContainerResponse](ci, "removeContainer", request)
}

func (ci *criImpl) ListContainers(request *ListContainersRequest) (*ListContainersResponse, error) {
	return criCallHelper[ListContainersRequest, ListContainersResponse](ci, "listContainers", request)
}

func (ci *criImpl) ContainerStatus(request *ContainerStatusRequest) (*ContainerStatusResponse, error) {
	return criCallHelper[ContainerStatusRequest, ContainerStatusResponse](ci, "containerStatus", request)
}

func (ci *criImpl) ListImages(request *ListImagesRequest) (*ListImagesResponse, error) {
	return criCallHelper[ListImagesRequest, ListImagesResponse](ci, "listImages", request)
}

func (ci *criImpl) PullImage(request *PullImageRequest) (*PullImageResponse, error) {
	return criCallHelper[PullImageRequest, PullImageResponse](ci, "pullImage", request)
}

func (ci *criImpl) RemoveImage(request *RemoveImageRequest) (*RemoveImageResponse, error) {
	return criCallHelper[RemoveImageRequest, RemoveImageResponse](ci, "removeImage", request)
}
