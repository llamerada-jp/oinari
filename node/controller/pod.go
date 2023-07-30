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
package controller

import (
	"encoding/json"
	"fmt"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/messaging/driver"
	"github.com/llamerada-jp/oinari/node/misc"
)

type ApplicationDigest struct {
	Name        string `json:"name"`
	Uuid        string `json:"uuid"`
	RunningNode string `json:"runningNode"`
	Owner       string `json:"owner"`
	State       string `json:"state"`
}

type PodController interface {
	DealLocalResource(raw []byte) error

	Create(name, owner, creatorNode string, spec *api.PodSpec) (*ApplicationDigest, error)
	GetPodData(uuid string) (*api.Pod, error)
	Migrate(uuid string, targetNodeID string) error
	Delete(uuid string) error
	Cleanup(uuid string) error
}

type podControllerImpl struct {
	podKvs    kvs.PodKvs
	messaging driver.MessagingDriver
	localNid  string
}

func NewPodController(podKvs kvs.PodKvs, messaging driver.MessagingDriver, localNid string) PodController {
	return &podControllerImpl{
		podKvs:    podKvs,
		messaging: messaging,
		localNid:  localNid,
	}
}

func (impl *podControllerImpl) DealLocalResource(raw []byte) error {
	pod := &api.Pod{}
	err := json.Unmarshal(raw, pod)
	if err != nil {
		return err
	}

	// check deletion
	if len(pod.Meta.DeletionTimestamp) != 0 {
		if len(pod.Status.RunningNode) == 0 || impl.getContainerStateDigest(pod) == api.ContainerStateTerminated {
			impl.podKvs.Delete(pod.Meta.Uuid)
			return nil
		}

		return impl.messaging.ReconcileContainer(pod.Status.RunningNode, pod.Meta.Uuid)
	}

	/*
		err = impl.accountMgr.BindPod(&pod)
		if err != nil {
			return err
		}
		//*/

	// waiting to schedule
	if len(pod.Status.RunningNode) == 0 {
		return impl.schedulePod(pod)
	}

	if pod.Status.RunningNode == pod.Status.TargetNode {
		stateDigest := impl.getContainerStateDigest(pod)
		if stateDigest == api.ContainerStateTerminated || stateDigest == api.ContainerStateUnknown {
			// TODO restart pod by the restart policy
			return nil
		}
	} else {
		stateDigest := impl.getContainerStateDigest(pod)
		if stateDigest == api.ContainerStateTerminated {
			pod.Status.RunningNode = pod.Status.TargetNode
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerStatus.ContainerID = ""
				containerStatus.Image = ""
				containerStatus.State = api.ContainerStateWaiting
			}
			return impl.podKvs.Update(pod)

		} else if stateDigest == api.ContainerStateUnknown {
			// TODO restart pod by the restart policy
		}
	}

	err = impl.messaging.ReconcileContainer(pod.Status.RunningNode, pod.Meta.Uuid)
	if err != nil {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStatus.State = api.ContainerStateUnknown
		}
		return impl.podKvs.Update(pod)
	}

	return nil
}

func (impl *podControllerImpl) Create(name, owner, creatorNode string, spec *api.PodSpec) (*ApplicationDigest, error) {
	pod := &api.Pod{
		Meta: &api.ObjectMeta{
			Type:        api.ResourceTypePod,
			Name:        name,
			Owner:       owner,
			CreatorNode: creatorNode,
			Uuid:        api.GeneratePodUuid(),
		},
		Spec: impl.setDefaultPodSpec(spec),
		Status: &api.PodStatus{
			ContainerStatuses: make([]api.ContainerStatus, 0),
		},
	}

	for _ = range pod.Spec.Containers {
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
			api.ContainerStatus{
				State: api.ContainerStateWaiting,
			})
	}

	err := impl.podKvs.Create(pod)
	// TODO: retry only if the same uuid id exists
	if err != nil {
		return nil, err
	}

	return &ApplicationDigest{
		Name:  name,
		Uuid:  pod.Meta.Uuid,
		Owner: pod.Meta.Owner,
		State: impl.getContainerStateMessage(pod),
	}, nil
}

func (impl *podControllerImpl) setDefaultPodSpec(spec *api.PodSpec) *api.PodSpec {
	if spec.Scheduler == nil {
		spec.Scheduler = &api.SchedulerSpec{
			Type: "creator",
		}
	}
	return spec
}

func (impl *podControllerImpl) schedulePod(pod *api.Pod) error {
	if len(pod.Status.RunningNode) != 0 {
		return nil
	}

	switch pod.Spec.Scheduler.Type {
	case "creator":
		pod.Status.RunningNode = pod.Meta.CreatorNode
		pod.Status.TargetNode = pod.Meta.CreatorNode
		return impl.podKvs.Update(pod)

	default:
		return fmt.Errorf("unsupported scheduling policy:%s", pod.Spec.Scheduler.Type)
	}
}

func (impl *podControllerImpl) getContainerStateDigest(pod *api.Pod) api.ContainerState {
	allTerminated := true
	hasRunning := false

	for _, containerState := range pod.Status.ContainerStatuses {
		switch containerState.State {
		case api.ContainerStateWaiting:
			allTerminated = false
		case api.ContainerStateRunning:
			allTerminated = false
			hasRunning = true
		case api.ContainerStateTerminated:
		case api.ContainerStateUnknown:
			return api.ContainerStateUnknown
		}
	}

	if allTerminated {
		return api.ContainerStateTerminated
	} else if hasRunning {
		return api.ContainerStateRunning
	} else {
		return api.ContainerStateWaiting
	}
}

func (impl *podControllerImpl) getContainerStateMessage(pod *api.Pod) string {
	digest := impl.getContainerStateDigest(pod)

	if digest == api.ContainerStateRunning {
		all := len(pod.Status.ContainerStatuses)
		running := 0

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State == api.ContainerStateRunning {
				running += 1
			}
		}

		return fmt.Sprintf("%s (%d/%d)", string(digest), running, all)
	}

	return string(digest)
}

func (impl *podControllerImpl) GetPodData(uuid string) (*api.Pod, error) {
	return impl.podKvs.Get(uuid)
}

func (impl *podControllerImpl) Migrate(uuid string, targetNodeID string) error {
	pod, err := impl.podKvs.Get(uuid)
	if err != nil {
		return err
	}

	if len(pod.Status.RunningNode) == 0 {
		pod.Status.RunningNode = targetNodeID
		pod.Status.TargetNode = targetNodeID

	} else {
		// TODO check if migration is accepted

		pod.Status.TargetNode = targetNodeID
	}

	return impl.podKvs.Update(pod)
}

func (impl *podControllerImpl) Delete(uuid string) error {
	pod, err := impl.podKvs.Get(uuid)
	if err != nil {
		return err
	}

	if len(pod.Meta.DeletionTimestamp) != 0 {
		pod.Meta.DeletionTimestamp = misc.GetTimestamp()
		return impl.podKvs.Update(pod)
	}

	return nil
}

func (impl *podControllerImpl) Cleanup(uuid string) error {
	pod, err := impl.podKvs.Get(uuid)
	if err != nil {
		return err
	}

	if impl.getContainerStateDigest(pod) != api.ContainerStateUnknown {
		return fmt.Errorf("target pod of cleanup should be unknown state")
	}

	return impl.podKvs.Delete(uuid)
}
