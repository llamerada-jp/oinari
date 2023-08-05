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
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/misc"
)

type ContainerController interface {
	GetLocalPodUUIDs() []string
	Reconcile(ctx context.Context, podUuid string) error
}

type reconcileState struct {
	running   bool
	sandboxID string
}

type containerControllerImpl struct {
	localNid string
	cri      cri.CRI
	podKvs   kvs.PodKvs
	// key: Pod UUID
	reconcileStates map[string]*reconcileState
	mtx             sync.Mutex
}

func NewContainerController(localNid string, cri cri.CRI, podKvs kvs.PodKvs) ContainerController {
	return &containerControllerImpl{
		localNid:        localNid,
		cri:             cri,
		podKvs:          podKvs,
		reconcileStates: make(map[string]*reconcileState),
	}
}

func (impl *containerControllerImpl) GetLocalPodUUIDs() []string {
	uuids := make([]string, len(impl.reconcileStates))
	for uuid := range impl.reconcileStates {
		uuids = append(uuids, uuid)
	}
	return uuids
}

func (impl *containerControllerImpl) Reconcile(ctx context.Context, podUuid string) error {
	state, running := func() (*reconcileState, bool) {
		impl.mtx.Lock()
		defer impl.mtx.Unlock()

		state, ok := impl.reconcileStates[podUuid]
		if !ok {
			state = &reconcileState{}
			impl.reconcileStates[podUuid] = state
		}

		running := state.running
		state.running = true

		return state, running
	}()

	// skip when other reconcile running
	if running {
		return nil
	}

	// set running flg and delete instance if required when end of reconcile
	deleteFlg := false
	defer func() {
		impl.mtx.Lock()
		defer impl.mtx.Unlock()

		state.running = false

		if deleteFlg || len(state.sandboxID) == 0 {
			delete(impl.reconcileStates, podUuid)
		}
	}()

	pod, err := impl.podKvs.Get(podUuid)
	if err != nil {
		return err
	}

	// force stop container if running node is not this node
	if pod.Status.RunningNode != impl.localNid {
		if len(state.sandboxID) != 0 {
			if err = impl.removeSandbox(state.sandboxID); err != nil {
				return err
			}
			deleteFlg = true
		}
		return nil
	}

	// terminate containers if deletion timestamp has set
	if len(pod.Meta.DeletionTimestamp) != 0 || pod.Status.TargetNode != pod.Status.RunningNode {
		if len(state.sandboxID) == 0 {
			return impl.updatePodInfo(state, pod)
		}

		terminated, err := impl.letTerminate(state.sandboxID)
		if err != nil {
			return err
		}

		if err = impl.updatePodInfo(state, pod); err != nil {
			return err
		}

		if terminated {
			// remove sandbox if containers have terminated
			if err = impl.removeSandbox(podUuid); err != nil {
				return err
			}

			deleteFlg = true
		}
		return nil
	}

	// make containers running as necessary
	err = impl.letRunning(state, pod)
	if err != nil {
		return err
	}

	return impl.updatePodInfo(state, pod)
}

// return sandboxId
func (impl *containerControllerImpl) letRunning(state *reconcileState, pod *api.Pod) error {
	// create sandbox if it isn't exist
	if len(state.sandboxID) == 0 {
		res, err := impl.cri.RunPodSandbox(&cri.RunPodSandboxRequest{
			Config: cri.PodSandboxConfig{
				Metadata: cri.PodSandboxMetadata{
					Name:      pod.Meta.Name,
					UID:       pod.Meta.Uuid,
					Namespace: "",
				},
			},
		})
		if err != nil {
			return err
		}

		state.sandboxID = res.PodSandboxId
	}

	// make containers as map[container name]ContainerStatus
	sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
		PodSandboxId: state.sandboxID,
	})
	if err != nil {
		return err
	}
	containers := make(map[string]cri.ContainerStatus)
	for _, containerStatus := range sandboxStatus.ContainersStatuses {
		containers[containerStatus.Metadata.Name] = containerStatus
	}

	// make images as map[image url]true
	imageList, err := impl.cri.ListImages(&cri.ListImagesRequest{})
	if err != nil {
		return err
	}
	images := make(map[string]bool)
	for _, image := range imageList.Images {
		images[image.Spec.Image] = true
	}

	// load image if necessary
	for idx, spec := range pod.Spec.Containers {
		status := pod.Status.ContainerStatuses[idx]
		if status.State.Running != nil {
			continue
		}

		_, containerExists := containers[spec.Name]
		if containerExists {
			continue
		}

		if _, imageExists := images[spec.Image]; !imageExists {
			_, err := impl.cri.PullImage(&cri.PullImageRequest{
				Image: cri.ImageSpec{
					Image: spec.Image,
				},
			})
			if err != nil {
				return err
			}
			images[spec.Image] = true
		}
	}

	for idx, spec := range pod.Spec.Containers {
		status := &pod.Status.ContainerStatuses[idx]
		_, containerExists := containers[spec.Name]

		// start containers if they are not exist
		if status.State.Running == nil && !containerExists {
			envs := []cri.KeyValue{}
			for _, one := range spec.Env {
				envs = append(envs, cri.KeyValue{
					Key:   one.Name,
					Value: one.Value,
				})
			}

			res, err := impl.cri.CreateContainer(&cri.CreateContainerRequest{
				PodSandboxId: state.sandboxID,
				Config: cri.ContainerConfig{
					Metadata: cri.ContainerMetadata{
						Name: spec.Name,
					},
					Image: cri.ImageSpec{
						Image: spec.Image,
					},
					Runtime: spec.Runtime,
					Args:    spec.Args,
					Envs:    envs,
				},
			})
			if err != nil {
				log.Printf("failed to create container: %w", err)
				continue
			}

			status.ContainerID = res.ContainerId
			status.State = api.ContainerState{
				Running: &api.ContainerStateRunning{
					StartedAt: misc.GetTimestamp(),
				},
			}

			_, err = impl.cri.StartContainer(&cri.StartContainerRequest{
				ContainerId: res.ContainerId,
			})
			if err != nil {
				log.Printf("failed to start container: %w", err)
				continue
			}
		}
	}

	return nil
}

func (impl *containerControllerImpl) letTerminate(sandboxId string) (bool, error) {
	// TODO: send exit signal when any container running

	// TODO: skip processing when all container exited

	// TODO: force exit all containers after timeout
}

func (impl *containerControllerImpl) updatePodInfo(state *reconcileState, pod *api.Pod) error {
	// make containers as map[container name]ContainerStatus
	containerStatuses := make(map[string]*cri.ContainerStatus)
	if len(state.sandboxID) != 0 {
		sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
			PodSandboxId: state.sandboxID,
		})
		if err != nil {
			return err
		}

		for _, containerStatus := range sandboxStatus.ContainersStatuses {
			containerStatuses[containerStatus.Metadata.Name] = &containerStatus
		}
	}

	for idx, spec := range pod.Spec.Containers {
		status := &pod.Status.ContainerStatuses[idx]
		container, containerExists := containerStatuses[spec.Name]

		if !containerExists {
			if status.State.Terminated == nil && status.State.Unknown == nil {
				status.State.Unknown = &api.ContainerStateUnknown{
					Timestamp: misc.GetTimestamp(),
					Reason:    "the container not found",
				}
			}
			continue
		}

		if (container.State == cri.ContainerRunning || container.State == cri.ContainerExited) && status.State.Running == nil {
			status.ContainerID = container.ID
			status.Image = container.Image.Image
			status.State.Running = &api.ContainerStateRunning{
				StartedAt: misc.GetTimestamp(),
			}
			status.State.Unknown = nil
		}

		if status.ContainerID != container.ID {
			impl.removeSandbox(state.sandboxID)
			if status.State.Unknown == nil {
				status.State.Unknown = &api.ContainerStateUnknown{
					Timestamp: misc.GetTimestamp(),
					Reason:    fmt.Sprintf("container id is different from actual (%s)", container.ID),
				}
			}
			continue
		}

		if container.State == cri.ContainerExited && status.State.Terminated == nil {
			status.State.Terminated = &api.ContainerStateTerminated{
				FinishedAt: container.FinishedAt,
				ExitCode:   container.ExitCode,
			}
			status.State.Unknown = nil
		}

		if status.State.Terminated == nil && status.State.Unknown == nil && container.State == cri.ContainerUnknown {
			status.State.Unknown = &api.ContainerStateUnknown{
				Timestamp: misc.GetTimestamp(),
				Reason:    "container status could not get",
			}
		}

		if status.State.Terminated != nil && container.State != cri.ContainerExited && container.State != cri.ContainerUnknown {
			impl.removeSandbox(state.sandboxID)
			return fmt.Errorf("container should be terminated")
		}

		delete(containerStatuses, spec.Name)
	}

	if err := impl.podKvs.Update(pod); err != nil {
		return err
	}

	if len(containerStatuses) > 0 {
		impl.removeSandbox(state.sandboxID)
		return fmt.Errorf("found differences in spec of pod between running containers")
	}

	return nil
}

func (impl *containerControllerImpl) removeSandbox(sandboxId string) error {
	cl, err := impl.cri.ListContainers(&cri.ListContainersRequest{
		Filter: &cri.ContainerFilter{
			PodSandboxId: sandboxId,
		},
	})
	if err != nil {
		return err
	}

	for _, container := range cl.Containers {
		_, err = impl.cri.RemoveContainer(&cri.RemoveContainerRequest{
			ContainerId: container.ID,
		})
		if err != nil {
			log.Printf("failed to remove container: %w", err)
		}
	}

	_, err = impl.cri.RemovePodSandbox(&cri.RemovePodSandboxRequest{
		PodSandboxId: sandboxId,
	})

	return err
}
