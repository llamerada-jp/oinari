package cri

/**
 * This interface is partial mimic of Kubernetes cri-api. And there are some differences
 * caused by Oinari using WASM on the web browsers. Oinari implements the interface
 * using crosslink internally without gRPC because it is difficult to implement it
 * using gRPC between Go(WASM) and TypeScript via web worker.
 * ref: https://github.com/kubernetes/cri-api
 */
type CRI interface {
	// apis for sandbox
	RunPodSandbox(*RunPodSandboxRequest) (*RunPodSandboxResponse, error)
	StopPodSandbox(*StopPodSandboxRequest) (*StopPodSandboxResponse, error)
	RemovePodSandbox(*RemovePodSandboxRequest) (*RemovePodSandboxResponse, error)
	PodSandboxStatus(*PodSandboxStatusRequest) (*PodSandboxStatusResponse, error)
	ListPodSandbox(*ListPodSandboxRequest) (*ListPodSandboxResponse, error)

	// apis for container
	CreateContainer(*CreateContainerRequest) (*CreateContainerResponse, error)
	StartContainer(*StartContainerRequest) (*StartContainerResponse, error)
	StopContainer(*StopContainerRequest) (*StopContainerResponse, error)
	RemoveContainer(*RemoveContainerRequest) (*RemoveContainerResponse, error)
	ListContainers(*ListContainersRequest) (*ListContainersResponse, error)
	ContainerStatus(*ContainerStatusRequest) (*ContainerStatusResponse, error)

	// apis for image
	ListImages(*ListImagesRequest) (*ListImagesResponse, error)
	PullImage(*PullImageRequest) (*PullImageResponse, error)
	RemoveImage(*RemoveImageRequest) (*RemoveImageResponse, error)
}

type RunPodSandboxRequest struct {
	Config PodSandboxConfig `json:"config"`
}

type PodSandboxConfig struct {
	Metadata PodSandboxMetadata `json:"metadata"`
}

type PodSandboxMetadata struct {
	Name string `json:"name"`
	// UID is equal to Pod UID in the Pod ObjectMeta. This value must be global unique in the Oinari system.
	UID       string `json:"uid"`
	Namespace string `json:"namespace"`
}

type RunPodSandboxResponse struct {
	// PodSandboxID is not equal to UID of PodSandboxMetadata. This value used in node local and unique in only the node.
	PodSandboxID string `json:"pod_sandbox_id"`
}

type StopPodSandboxRequest struct {
	PodSandboxID string `json:"pod_sandbox_id"`
}

type StopPodSandboxResponse struct {
	// empty
}

type RemovePodSandboxRequest struct {
	PodSandboxID string `json:"pod_sandbox_id"`
}

type RemovePodSandboxResponse struct {
	// empty
}

type PodSandboxStatusRequest struct {
	PodSandboxID string `json:"pod_sandbox_id"`
}

type PodSandboxStatusResponse struct {
	Status             PodSandboxStatus  `json:"status"`
	ContainersStatuses []ContainerStatus `json:"containers_statuses"`
	Timestamp          string            `json:"timestamp"`
}

type ListPodSandboxRequest struct {
	Filter *PodSandboxFilter `json:"filter"`
}

type ListPodSandboxResponse struct {
	Items []PodSandbox `json:"items"`
}

type PodSandboxStatus struct {
	ID        string             `json:"id"`
	Metadata  PodSandboxMetadata `json:"metadata"`
	State     PodSandboxState    `json:"state"`
	CreatedAt string             `json:"created_at"`
}

type PodSandboxState int

const (
	SandboxReady PodSandboxState = iota
	SandboxNotReady
)

type PodSandboxFilter struct {
	ID    string                `json:"id"`
	State *PodSandboxStateValue `json:"state"`
	// LabelSelector map[string]string     `json:"label_selector"`
}

type PodSandboxStateValue struct {
	State PodSandboxState `json:"state"`
}

type PodSandbox struct {
	ID        string             `json:"ID"`
	Metadata  PodSandboxMetadata `json:"metadata"`
	State     PodSandboxState    `json:"state"`
	CreatedAt string             `json:"created_at"`
	// Labels      map[string]string  `json:"labels"`
	// Annotations map[string]string  `json:"annotations"`
	// RuntimeHandler string          `json:"runtime_handler`
}

type CreateContainerRequest struct {
	PodSandboxID string          `json:"pod_sandbox_id"`
	Config       ContainerConfig `json:"config"`
	// SandboxConfig PodSandboxConfig `json:"sandbox_config"`
}

type ContainerConfig struct {
	Metadata ContainerMetadata `json:"metadata"`
	Image    ImageSpec         `json:"image"`
	Args     []string          `json:"args"`
	Envs     []KeyValue        `json:"envs"`
}

type ContainerMetadata struct {
	Name string `json:"name"`
}

type CreateContainerResponse struct {
	ContainerID string `json:"container_id"`
}

type StartContainerRequest struct {
	ContainerID string `json:"container_id"`
}

type StartContainerResponse struct {
	// empty
}

type StopContainerRequest struct {
	ContainerID string `json:"container_id"`
}

type StopContainerResponse struct {
	// empty
}

type RemoveContainerRequest struct {
	ContainerID string `json:"container_id"`
}

type RemoveContainerResponse struct {
	// empty
}

type ListContainersRequest struct {
	Filter *ContainerFilter `json:"filter"`
}

type ListContainersResponse struct {
	Containers []Container `json:"containers"`
}

type ContainerStatusRequest struct {
	ContainerID string `json:"container_id"`
	// Verbose bool `json:"verbose"`
}

type ContainerStatusResponse struct {
	Status ContainerStatus `json:"status"`
	// Info map[string]string `json:"info"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ContainerStatus struct {
	ID         string            `json:"id"`
	Metadata   ContainerMetadata `json:"metadata"`
	State      ContainerState    `json:"state"`
	CreatedAt  string            `json:"created_at"`
	StartedAt  string            `json:"started_at"`
	FinishedAt string            `json:"finished_at"`
	ExitCode   int               `json:"exit_code"`
	Image      ImageSpec         `json:"image"`
	ImageRef   string            `json:"image_ref"`
}

type ContainerState int

const (
	ContainerCreated ContainerState = iota
	ContainerRunning
	ContainerExited
	ContainerUnknown
)

type ContainerFilter struct {
	ID           string               `json:"id"`
	State        *ContainerStateValue `json:"state"`
	PodSandboxID string               `json:"pod_sandbox_id"`
}

type ContainerStateValue struct {
	State ContainerState `json:"state"`
}

type Container struct {
	ID           string            `json:"id"`
	PodSandboxID string            `json:"pod_sandbox_id"`
	Metadata     ContainerMetadata `json:"metadata"`
	Image        ImageSpec         `json:"image"`
	ImageRef     string            `json:"image_ref"`
	State        ContainerState    `json:"state"`
	CreatedAt    string            `json:"created_at"`
}

type ListImagesRequest struct {
	Filter ImageFilter `json:"filter"`
}

type ImageFilter struct {
	Image ImageSpec `json:"image"`
}

type ImageSpec struct {
	Image string `json:"image"`
}

type ListImagesResponse struct {
	Images []Image `json:"images"`
}

type Image struct {
	ID   string    `json:"id"`
	Spec ImageSpec `json:"spec"`
	// this field meaning the runtime environment of wasm, like  'core:dev1', 'go:1.19'
	Runtime []string `json:"runtime"`
}

type PullImageRequest struct {
	Image ImageSpec `json:"image"`
}

type PullImageResponse struct {
	ImageRef string `json:"image_ref"`
}

type RemoveImageRequest struct {
	Image ImageSpec `json:"image"`
}

type RemoveImageResponse struct {
	// nothing
}
