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

	"github.com/llamerada-jp/oinari/api/core"
	coreAPI "github.com/llamerada-jp/oinari/node/apis/core"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/misc"
)

const (
	ContainerLabelPodUUID = "pod-uuid"
)

type ContainerController interface {
	GetContainerInfos() []*ContainerInfo
	Reconcile(ctx context.Context, podUuid string) error
}

type ContainerInfo struct {
	PodUUID   string
	Owner     string
	SandboxID string
}

type reconcileState struct {
	// true if reconcile running
	running bool
	// will delete when reconcile finished
	willDelete    bool
	containerInfo ContainerInfo
}

type containerControllerImpl struct {
	localNid             string
	cri                  cri.CRI
	appFilter            ApplicationFilter
	podKvs               kvs.PodKvs
	recordKvs            kvs.RecordKvs
	apiCoreDriverManager *coreAPI.Manager
	// key: Pod UUID
	reconcileStates map[string]*reconcileState
	mtx             sync.Mutex
}

func NewContainerController(localNid string, cri cri.CRI, appFilter ApplicationFilter, podKvs kvs.PodKvs, recordKVS kvs.RecordKvs, apiCoreDriverManager *coreAPI.Manager) ContainerController {
	return &containerControllerImpl{
		localNid:             localNid,
		cri:                  cri,
		appFilter:            appFilter,
		podKvs:               podKvs,
		recordKvs:            recordKVS,
		apiCoreDriverManager: apiCoreDriverManager,
		reconcileStates:      make(map[string]*reconcileState),
	}
}

func (impl *containerControllerImpl) GetContainerInfos() []*ContainerInfo {
	impl.mtx.Lock()
	defer impl.mtx.Unlock()

	infos := make([]*ContainerInfo, 0)
	for _, state := range impl.reconcileStates {
		infos = append(infos, &state.containerInfo)
	}
	return infos
}

func (impl *containerControllerImpl) Reconcile(ctx context.Context, podUUID string) error {
	state, running := func() (*reconcileState, bool) {
		impl.mtx.Lock()
		defer impl.mtx.Unlock()

		state, ok := impl.reconcileStates[podUUID]
		if !ok {
			state = &reconcileState{
				containerInfo: ContainerInfo{
					PodUUID: podUUID,
				},
			}
			impl.reconcileStates[podUUID] = state
		}

		running := state.running
		state.running = true

		return state, running
	}()

	// skip when other reconcile running
	if running {
		return nil
	}

	defer func() {
		impl.mtx.Lock()
		defer impl.mtx.Unlock()

		state.running = false

		if state.willDelete {
			delete(impl.reconcileStates, podUUID)
		}
	}()

	pod, err := impl.podKvs.Get(podUUID)
	if err != nil {
		return err
	}

	state.containerInfo.Owner = pod.Meta.Owner

	// force stop container if running node is not this node
	if pod.Status.RunningNode != impl.localNid {
		if len(state.containerInfo.SandboxID) != 0 {
			if err = impl.removeSandbox(state.containerInfo.SandboxID); err != nil {
				return err
			}
			state.willDelete = true
		}
		return nil
	}

	// terminate containers if deletion timestamp has set
	if len(pod.Meta.DeletionTimestamp) != 0 || (len(pod.Spec.TargetNode) != 0 && pod.Spec.TargetNode != pod.Status.RunningNode) {
		if len(state.containerInfo.SandboxID) == 0 {
			return impl.updatePodInfo(state, pod)
		}

		if err := impl.letTerminate(state, pod); err != nil {
			return err
		}

		if err = impl.updatePodInfo(state, pod); err != nil {
			return err
		}

		terminated := true
		for _, containerStates := range pod.Status.ContainerStatuses {
			if containerStates.State.Terminated == nil {
				terminated = false
				break
			}
		}

		if terminated {
			// remove sandbox if containers have terminated
			if err = impl.removeSandbox(state.containerInfo.SandboxID); err != nil {
				return err
			}

			state.willDelete = true
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
func (impl *containerControllerImpl) letRunning(state *reconcileState, pod *core.Pod) error {
	if !impl.appFilter.IsAllowed(pod) {
		log.Printf("the application is not allowed to run on this node: %s/%s", pod.Meta.Owner, pod.Meta.Name)
		return nil
	}

	// create sandbox if it isn't exist
	if len(state.containerInfo.SandboxID) == 0 {
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
			state.willDelete = true
			return err
		}

		state.containerInfo.SandboxID = res.PodSandboxId
	}

	// make containers as map[container name]ContainerStatus
	sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
		PodSandboxId: state.containerInfo.SandboxID,
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
		if status.State.Running == nil {
			var containerID string
			if !containerExists {
				envs := []cri.KeyValue{}
				for _, one := range spec.Env {
					envs = append(envs, cri.KeyValue{
						Key:   one.Name,
						Value: one.Value,
					})
				}

				res, err := impl.cri.CreateContainer(&cri.CreateContainerRequest{
					PodSandboxId: state.containerInfo.SandboxID,
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
						Labels: map[string]string{
							ContainerLabelPodUUID: pod.Meta.Uuid,
						},
					},
				})
				if err != nil {
					log.Printf("failed to create container: %s", err.Error())
					continue
				}

				// create api driver
				impl.apiCoreDriverManager.NewCoreDriver(res.ContainerId, spec.Runtime)

				containerID = res.ContainerId
			}

			containers, err := impl.cri.ListContainers(&cri.ListContainersRequest{
				Filter: &cri.ContainerFilter{
					ID: containerID,
				},
			})
			if err != nil || len(containers.Containers) == 0 {
				log.Printf("failed to get container info: %s", err.Error())
				continue
			}

			status.ContainerID = containerID
			status.Image = containers.Containers[0].Image.Image

			if containers.Containers[0].State != cri.ContainerRunning && containers.Containers[0].State != cri.ContainerExited {
				_, err = impl.cri.StartContainer(&cri.StartContainerRequest{
					ContainerId: containerID,
				})
				if err != nil {
					log.Printf("failed to start container: %s", err.Error())
					continue
				}
			}

			status.State = core.ContainerState{
				Running: &core.ContainerStateRunning{
					StartedAt: misc.GetTimestamp(),
				},
			}
		}
	}

	return nil
}

func (impl *containerControllerImpl) letTerminate(state *reconcileState, pod *core.Pod) error {
	// TODO: send exit signal when any container running
	containers, err := impl.cri.ListContainers(&cri.ListContainersRequest{
		Filter: &cri.ContainerFilter{
			PodSandboxId: state.containerInfo.SandboxID,
		},
	})
	if err != nil {
		return err
	}

	isFinalize := len(pod.Meta.DeletionTimestamp) != 0
	var record *core.Record
	if !isFinalize {
		var err error
		record, err = impl.recordKvs.Get(pod.Meta.Uuid)
		if err != nil {
			return err
		}
		if record == nil {
			record = &core.Record{
				Meta: &core.ObjectMeta{
					Type:        core.ResourceTypeRecord,
					Name:        pod.Meta.Name,
					Owner:       pod.Meta.Owner,
					CreatorNode: pod.Meta.CreatorNode,
					Uuid:        pod.Meta.Uuid,
				},
				Data: &core.RecordData{
					Entries: make(map[string]core.RecordEntry),
				},
			}
		}
	}

	for _, container := range containers.Containers {
		if container.State == cri.ContainerExited {
			continue
		}

		raw, err := impl.apiCoreDriverManager.GetDriver(container.ID).Teardown(isFinalize)
		if err != nil {
			return fmt.Errorf("failed to teardown container: %w", err)
		}
		if raw != nil {
			record.Data.Entries[container.Metadata.Name] = core.RecordEntry{
				Record:    raw,
				Timestamp: misc.GetTimestamp(),
			}
		}

		_, err = impl.cri.StopContainer(&cri.StopContainerRequest{
			ContainerId: container.ID,
		})
		if err != nil {
			log.Printf("failed to stop container :%s", err.Error())
		}

		impl.apiCoreDriverManager.DestroyDriver(container.ID)
	}

	if !isFinalize {
		err = impl.recordKvs.Set(record)
		if err != nil {
			return err
		}
	} else {
		err = impl.recordKvs.Delete(pod.Meta.Uuid)
		if err != nil {
			log.Printf("failed to delete record: %s", err.Error())
		}
	}

	// TODO: skip processing when all container exited
	// TODO: force exit all containers after timeout

	return nil
}

func (impl *containerControllerImpl) updatePodInfo(state *reconcileState, pod *core.Pod) error {
	// make containers as map[container name]ContainerStatus
	containerStatuses := make(map[string]*cri.ContainerStatus)
	if len(state.containerInfo.SandboxID) != 0 {
		sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
			PodSandboxId: state.containerInfo.SandboxID,
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
				status.State.Unknown = &core.ContainerStateUnknown{
					Timestamp: misc.GetTimestamp(),
					Reason:    "the container not found",
				}
			}
			continue
		}

		if (container.State == cri.ContainerRunning || container.State == cri.ContainerExited) && status.State.Running == nil {
			status.ContainerID = container.ID
			status.Image = container.Image.Image
			status.State.Running = &core.ContainerStateRunning{
				StartedAt: misc.GetTimestamp(),
			}
			status.State.Unknown = nil
		}

		if status.ContainerID != container.ID {
			impl.removeSandbox(state.containerInfo.SandboxID)
			if status.State.Unknown == nil {
				status.State.Unknown = &core.ContainerStateUnknown{
					Timestamp: misc.GetTimestamp(),
					Reason:    fmt.Sprintf("container id is different from actual (%s)", container.ID),
				}
			}
			continue
		}

		if container.State == cri.ContainerExited && status.State.Terminated == nil {
			if status.State.Terminated != nil {
				status.LastState = status.State.Terminated
			}
			status.State.Terminated = &core.ContainerStateTerminated{
				FinishedAt: container.FinishedAt,
				ExitCode:   container.ExitCode,
			}
			status.State.Unknown = nil
		}

		if status.State.Terminated == nil && status.State.Unknown == nil && container.State == cri.ContainerUnknown {
			status.State.Unknown = &core.ContainerStateUnknown{
				Timestamp: misc.GetTimestamp(),
				Reason:    "container status could not get",
			}
		}

		if status.State.Terminated != nil && container.State != cri.ContainerExited && container.State != cri.ContainerUnknown {
			impl.removeSandbox(state.containerInfo.SandboxID)
			return fmt.Errorf("container should be terminated")
		}

		delete(containerStatuses, spec.Name)
	}

	if err := impl.podKvs.Update(pod); err != nil {
		return fmt.Errorf("failed to update pod info: %w", err)
	}

	if len(containerStatuses) > 0 {
		impl.removeSandbox(state.containerInfo.SandboxID)
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
			log.Printf("failed to remove container: %s", err.Error())
		}
	}

	_, err = impl.cri.RemovePodSandbox(&cri.RemovePodSandboxRequest{
		PodSandboxId: sandboxId,
	})

	return err
}
