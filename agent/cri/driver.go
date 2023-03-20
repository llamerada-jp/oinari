package cri

import (
	"encoding/json"
	"strings"

	"github.com/llamerada-jp/oinari/agent/crosslink"
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

	reqJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	ch := make(chan *RES)
	var funcError error

	ci.cl.Call(string(reqJson), map[string]string{
		crosslink.TAG_PATH: strings.Join([]string{crosslinkPath, path}, "/"),
	}, func(result string, err error) {
		defer close(ch)

		if err != nil {
			funcError = err
			return
		}

		var res RES
		err = json.Unmarshal([]byte(result), &res)
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
