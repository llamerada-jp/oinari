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
