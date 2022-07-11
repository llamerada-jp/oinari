package core

type ResourceType string

const (
	ResourceTypePod = ResourceType("pod")
)

type ObjectMeta struct {
	Type ResourceType `json:"type"`
	Name string       `json:"name"`
	Uuid string       `json:"uuid"`
}

type Pod struct {
	Meta   ObjectMeta `json:"meta"`
	Spec   PodSpec    `json:"spec"`
	Status PodStatus  `json:"status"`
}

type PodSpec struct {
	CreatorNode string          `json:"creatorNode"`
	Containers  []ContainerSpec `json:"containers"`
	Scheduler   SchedulerSpec   `json:"scheduler"`
}

type ContainerSpec struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type SchedulerSpec struct {
	Type string `json:"type"`
}

type PodStatus struct {
	RunningNode string `json:"runningNode"`
}
