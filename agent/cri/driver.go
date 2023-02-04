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
	type ResErr struct {
		res *RES
		err error
	}

	reqJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	ch := make(chan *ResErr)

	ci.cl.Call(string(reqJson), map[string]string{
		crosslink.TAG_PATH: strings.Join([]string{crosslinkPath, path}, "/"),
	}, func(result string, err error) {
		if err != nil {
			ch <- &ResErr{nil, err}
			return
		}

		var res RES
		err = json.Unmarshal([]byte(result), &res)
		if err != nil {
			ch <- &ResErr{nil, err}
			return
		}

		ch <- &ResErr{&res, nil}
	})

	resErr := <-ch
	return resErr.res, resErr.err
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

func (ci *criImpl) ListImages(request *ListImagesRequest) (*ListImagesResponse, error) {
	return criCallHelper[ListImagesRequest, ListImagesResponse](ci, "listImages", request)
}

func (ci *criImpl) PullImage(request *PullImageRequest) (*PullImageResponse, error) {
	return criCallHelper[PullImageRequest, PullImageResponse](ci, "pullImage", request)
}

func (ci *criImpl) RemoveImage(request *RemoveImageRequest) (*RemoveImageResponse, error) {
	return criCallHelper[RemoveImageRequest, RemoveImageResponse](ci, "removeImage", request)
}