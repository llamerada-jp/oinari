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
	"strings"

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
		if len(pod.Status.RunningNode) == 0 || impl.checkContainerStateTerminated(pod) {
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
		if impl.checkContainerStateTerminated(pod) || impl.checkContainerStateUnknown(pod) {
			// TODO restart pod by the restart policy
			return nil
		}

	} else {
		if impl.checkContainerStateTerminated(pod) {
			pod.Status.RunningNode = pod.Status.TargetNode
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerStatus.ContainerID = ""
				containerStatus.Image = ""
				containerStatus.State = api.ContainerState{}
			}
			return impl.podKvs.Update(pod)

		} else if impl.checkContainerStateUnknown(pod) {
			// TODO restart pod by the restart policy
		}
	}

	err = impl.messaging.ReconcileContainer(pod.Status.RunningNode, pod.Meta.Uuid)
	if err != nil {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStatus.State.Unknown = &api.ContainerStateUnknown{
				Timestamp: misc.GetTimestamp(),
				Reason:    fmt.Sprintf("failed to call reconciliation to %s: %w", pod.Status.RunningNode, err),
			}
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
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, api.ContainerStatus{})
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

func (impl *podControllerImpl) getContainerStateMessage(pod *api.Pod) string {
	waiting := 0
	running := 0
	terminated := 0
	unknownReasons := make([]string, 0)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Terminated != nil {
			terminated += 1

		} else if containerStatus.State.Unknown != nil {
			unknownReasons = append(unknownReasons, containerStatus.State.Unknown.Reason)

		} else if containerStatus.State.Running != nil {
			running += 1

		} else {
			waiting += 1
		}
	}

	message := fmt.Sprintf("waiting:%d / running:%d / terminated:%d / unknown:%d", waiting, running, terminated, len(unknownReasons))
	if len(unknownReasons) != 0 {
		message = fmt.Sprintf("%s\n%s", message, strings.Join(unknownReasons, "\n"))
	}
	return message
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

	if impl.checkContainerStateUnknown(pod) {
		return fmt.Errorf("target pod of cleanup should be unknown state")
	}

	return impl.podKvs.Delete(uuid)
}

func (impl *podControllerImpl) checkContainerStateTerminated(pod *api.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Terminated == nil {
			return false
		}
	}

	return true
}

func (impl *podControllerImpl) checkContainerStateUnknown(pod *api.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Terminated == nil && containerStatus.State.Unknown != nil {
			return true
		}
	}

	return false
}
