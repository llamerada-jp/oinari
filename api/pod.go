package api

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
