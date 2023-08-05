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

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/kvs"
)

type ContainerController interface {
	GetLocalPodUUIDs() []string
	Reconcile(ctx context.Context, podUuid string) error
}

type containerControllerImpl struct {
	localNid string
	cri      cri.CRI
	podKvs   kvs.PodKvs
	// key: Pod UUID, value: sandbox ID
	sandboxIdMap map[string]string
}

func NewContainerController(localNid string, cri cri.CRI, podKvs kvs.PodKvs) ContainerController {
	return &containerControllerImpl{
		localNid:     localNid,
		cri:          cri,
		podKvs:       podKvs,
		sandboxIdMap: make(map[string]string),
	}
}

func (impl *containerControllerImpl) GetLocalPodUUIDs() []string {
	uuids := make([]string, len(impl.sandboxIdMap))
	for uuid := range impl.sandboxIdMap {
		uuids = append(uuids, uuid)
	}
	return uuids
}

func (impl *containerControllerImpl) Reconcile(ctx context.Context, podUuid string) error {
	pod, err := impl.podKvs.Get(podUuid)
	if err != nil {
		return err
	}

	// force stop container if running node is not this node
	if pod.Status.RunningNode != impl.localNid {
		return impl.removeSandbox(podUuid)
	}

	// terminate containers if deletion timestamp has set
	if len(pod.Meta.DeletionTimestamp) != 0 {
		sandboxId, exists := impl.sandboxIdMap[podUuid]
		if !exists {
			return impl.updatePodInfoDelete(pod, "")
		}

		terminated, err := impl.letTerminate(sandboxId)
		if err != nil {
			return err
		}

		// remove sandbox if containers have terminated
		if terminated {
			if err = impl.removeSandbox(podUuid); err != nil {
				return err
			}
		}

		if err = impl.updatePodInfoDelete(pod, sandboxId); err != nil {
			return err
		}

		delete(impl.sandboxIdMap, podUuid)
		return nil
	}

	// make containers running as necessary
	sandboxId, err := impl.letRunning(pod)
	if err != nil {
		return err
	}

	return impl.updatePodInfoRunning(pod, sandboxId)
}

// return sandboxId
func (impl *containerControllerImpl) letRunning(pod *api.Pod) (string, error) {
	sandboxId, sandboxExists := impl.sandboxIdMap[pod.Meta.Uuid]

	// create sandbox if it isn't exist
	if !sandboxExists {
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
			return "", err
		}

		sandboxId = res.PodSandboxId
		impl.sandboxIdMap[pod.Meta.Uuid] = sandboxId
	}

	// make containers as map[container name]ContainerStatus
	sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
		PodSandboxId: sandboxId,
	})
	if err != nil {
		return "", err
	}
	containers := make(map[string]cri.ContainerStatus)
	for _, containerStatus := range sandboxStatus.ContainersStatuses {
		containers[containerStatus.Metadata.Name] = containerStatus
	}

	// make images as map[image url]true
	imageList, err := impl.cri.ListImages(&cri.ListImagesRequest{})
	if err != nil {
		return "", err
	}
	images := make(map[string]bool)
	for _, image := range imageList.Images {
		images[image.Spec.Image] = true
	}

	// load image if necessary
	for idx, spec := range pod.Spec.Containers {
		status := pod.Status.ContainerStatuses[idx]
		if status.State != api.ContainerStateWaiting {
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
				return "", err
			}
			images[spec.Image] = true
		}
	}

	for idx, spec := range pod.Spec.Containers {
		status := pod.Status.ContainerStatuses[idx]
		_, containerExists := containers[spec.Name]

		// start containers if they are not exist
		if status.State == api.ContainerStateWaiting && !containerExists {
			envs := []cri.KeyValue{}
			for _, one := range spec.Env {
				envs = append(envs, cri.KeyValue{
					Key:   one.Name,
					Value: one.Value,
				})
			}

			res, err := impl.cri.CreateContainer(&cri.CreateContainerRequest{
				PodSandboxId: sandboxId,
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
			status.State = api.ContainerStateRunning

			_, err = impl.cri.StartContainer(&cri.StartContainerRequest{
				ContainerId: res.ContainerId,
			})
			if err != nil {
				log.Printf("failed to start container: %w", err)
				continue
			}
		}
	}

	return sandboxId, nil
}

func (impl *containerControllerImpl) updatePodInfoRunning(pod *api.Pod, sandboxId string) error {
	// make containers as map[container name]ContainerStatus
	sandboxStatus, err := impl.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
		PodSandboxId: sandboxId,
	})
	if err != nil {
		return err
	}

	containerStatuses := make(map[string]*cri.ContainerStatus)
	for _, containerStatus := range sandboxStatus.ContainersStatuses {
		containerStatuses[containerStatus.Metadata.Name] = &containerStatus
	}

	for idx, spec := range pod.Spec.Containers {
		status := pod.Status.ContainerStatuses[idx]
		container, containerExists := containerStatuses[spec.Name]

		if status.ContainerID != container.ID {
			impl.removeSandbox(sandboxId)
			return fmt.Errorf("found a wrong container id")
		}

		if status.State == api.ContainerStateTerminated {
			if container.State != cri.ContainerExited {
				impl.removeSandbox(sandboxId)
				return fmt.Errorf("container should be terminated")
			}
			continue
		}

		if !containerExists {
			status.State = api.ContainerStateUnknown
		}

		switch container.State {
		case cri.ContainerCreated:
			status.State = api.ContainerStateRunning
		case cri.ContainerRunning:
			status.State = api.ContainerStateRunning
		case cri.ContainerExited:
			status.State = api.ContainerStateTerminated
		case cri.ContainerUnknown:
			status.State = api.ContainerStateUnknown
		}

		delete(containerStatuses, spec.Name)
	}

	if err = impl.podKvs.Update(pod); err != nil {
		return err
	}

	if len(containerStatuses) > 0 {
		impl.removeSandbox(sandboxId)
		return fmt.Errorf("found differences in spec of pod between running containers")
	}

	return nil
}

func (impl *containerControllerImpl) letTerminate(podUuid string) (bool, error) {
	// TODO: send exit signal when any container running
	// TODO: skip processing when all container exited
	// TODO: force exit all containers after timeout
}

func (impl *containerControllerImpl) updatePodInfoDelete(pod *api.Pod, sandboxId string) error {

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
