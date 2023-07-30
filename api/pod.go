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
package api

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

var ContainerRuntimeRequired = []string{
	"go:1.19",
	"go:1.20",
}
var ContainerRuntimeAccepted = []string{}

type Pod struct {
	Meta   *ObjectMeta `json:"meta"`
	Spec   *PodSpec    `json:"spec"`
	Status *PodStatus  `json:"status"`
}

type PodSpec struct {
	Containers    []ContainerSpec `json:"containers"`
	Scheduler     *SchedulerSpec  `json:"scheduler"`
	EnableMigrate bool            `json:"enableMigrate"`
}

type RestartPolicy string

const (
	RestartPolicyDisable         RestartPolicy = "Disable"
	RestartPolicyAlways          RestartPolicy = "Always"
	RestartPolicyStrictExited    RestartPolicy = "StrictExited"
	RestartPolicyStrictSucceeded RestartPolicy = "StrictSucceeded"
	RestartPolicyStrictFailed    RestartPolicy = "StrictFailed"
	RestartPolicyOnce            RestartPolicy = "Once"
)

var RestartPolicyAccepted = []RestartPolicy{
	RestartPolicyDisable,
	RestartPolicyAlways,
	RestartPolicyStrictExited,
	RestartPolicyStrictSucceeded,
	RestartPolicyStrictFailed,
	RestartPolicyOnce,
}

type ContainerSpec struct {
	Name          string        `json:"name"`
	Image         string        `json:"image"`
	Runtime       []string      `json:"runtime"`
	Args          []string      `json:"args"`
	Env           []EnvVar      `json:"env"`
	RestartPolicy RestartPolicy `json:"restartPolicy"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	// valueFrom is not supported yet.
}

type SchedulerSpec struct {
	Type string `json:"type"`
}

type ContainerState string

const (
	ContainerStateWaiting    ContainerState = "Waiting"
	ContainerStateRunning    ContainerState = "Running"
	ContainerStateTerminated ContainerState = "Terminated"
	ContainerStateUnknown    ContainerState = "Unknown"
)

var ContainerStateAccepted = []ContainerState{
	ContainerStateWaiting,
	ContainerStateRunning,
	ContainerStateTerminated,
	ContainerStateUnknown,
}

type ContainerStatus struct {
	ContainerID string         `json:"containerID"`
	Image       string         `json:"image"`
	State       ContainerState `json:"state"`
}

type PodStatus struct {
	RunningNode       string            `json:"runningNode"`
	TargetNode        string            `json:"targetNode"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses"`
}

func GeneratePodUuid() string {
	return uuid.Must(uuid.NewRandom()).String()
}

func ValidatePodUuid(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid (%s): %w", id, err)
	}
	return nil
}

func (pod *Pod) Validate(mustStatus bool) error {
	if err := pod.Meta.Validate(ResourceTypePod); err != nil {
		return fmt.Errorf("invalid meta field: %w", err)
	}

	if err := ValidatePodUuid(pod.Meta.Uuid); err != nil {
		return fmt.Errorf("invalid uuid in pod meta field: %w", err)
	}

	if err := pod.Spec.validate(); err != nil {
		return err
	}

	// skip checking status if is not required or not set
	if !mustStatus && pod.Status == nil {
		return nil
	}

	if pod.Status == nil {
		return fmt.Errorf("pod status should be filled")
	}

	if err := pod.Status.validate(len(pod.Spec.Containers)); err != nil {
		return err
	}

	return nil
}

func (spec *PodSpec) validate() error {
	if spec == nil {
		return fmt.Errorf("pod spec should be filled")
	}

	if len(spec.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	containerNames := []string{}

	for _, container := range spec.Containers {
		// Name filed
		if len(container.Name) == 0 {
			return fmt.Errorf("name of the container should be specify")
		}

		if slices.Contains(containerNames, container.Name) {
			return fmt.Errorf("name of the container should be unique in the pod")
		}
		containerNames = append(containerNames, container.Name)

		// Image field
		if len(container.Image) == 0 {
			return fmt.Errorf("image of the container should be specify")
		}

		u, err := url.ParseRequestURI(container.Image)
		if err != nil {
			return fmt.Errorf("image of the container should be URI formatted")
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("image of the container should be http or https scheme")
		}

		// Runtime field
		if len(container.Runtime) == 0 {
			return fmt.Errorf("at least on runtime is required for the container")
		}

		count := 0
		for _, r := range container.Runtime {
			if slices.Contains(ContainerRuntimeRequired, r) {
				count += 1

			} else if !slices.Contains(ContainerRuntimeAccepted, r) {
				return fmt.Errorf("there is an unsupported runtime in the container")
			}
		}
		if count != 1 {
			return fmt.Errorf("there should be just one required runtime in the container")
		}

		// Env field
		envNames := []string{}
		for _, e := range container.Env {
			if slices.Contains(envNames, e.Name) {
				return fmt.Errorf("env name should be unique in the container")
			}
			envNames = append(envNames, e.Name)
		}

		// RestartPolicy field
		if !slices.Contains(RestartPolicyAccepted, container.RestartPolicy) {
			return fmt.Errorf("there is an unsupported restart policy in the container")
		}
	}

	return nil
}

func (status *PodStatus) validate(containerNum int) error {
	// RunningNode and TargetNode field
	if len(status.RunningNode) != 0 && ValidateNodeId(status.RunningNode) != nil {
		return fmt.Errorf("invalid running node id specified in the pod status")
	}

	if len(status.TargetNode) != 0 && ValidateNodeId(status.TargetNode) != nil {
		return fmt.Errorf("invalid target node id specified in the pod status")
	}

	if (len(status.RunningNode) == 0) != (len(status.TargetNode) == 0) {
		return fmt.Errorf("running node and target node should be specified or both should be empty")
	}

	// ContainerStatuses field
	if len(status.ContainerStatuses) != containerNum {
		return fmt.Errorf("container statues count should be equal to the containers in the spec field")
	}

	for _, containerState := range status.ContainerStatuses {
		// State field
		if !slices.Contains(ContainerStateAccepted, containerState.State) {
			return fmt.Errorf("there is an unsupported container state in the pod status")
		}
	}

	return nil
}
