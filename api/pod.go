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

	"github.com/google/uuid"
)

type Pod struct {
	Meta   *ObjectMeta `json:"meta"`
	Spec   *PodSpec    `json:"spec"`
	Status *PodStatus  `json:"status"`
}

type PodSpec struct {
	Containers []ContainerSpec `json:"containers"`
	Scheduler  *SchedulerSpec  `json:"scheduler"`
}

type ContainerSpec struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type SchedulerSpec struct {
	Type string `json:"type"`
}

type PodPhase string

const (
	PodPhasePending   PodPhase = "Pending"
	PodPhaseRunning   PodPhase = "Running"
	PodPhaseMigrating PodPhase = "Migrating"
	PodPhaseExited    PodPhase = "Exited"
)

type PodStatus struct {
	RunningNode string   `json:"runningNode"`
	Phase       PodPhase `json:"phase"`
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
