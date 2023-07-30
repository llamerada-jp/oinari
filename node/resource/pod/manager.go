package pod

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/misc"
)

type ApplicationDigest struct {
	Name        string `json:"name"`
	Uuid        string `json:"uuid"`
	RunningNode string `json:"runningNode"`
	Owner       string `json:"owner"`
	State       string `json:"state"`
}

type Manager interface {
	DealLocalResource(raw []byte) error
	Loop(ctx context.Context) error

	Create(name, owner, creatorNode string, spec *api.PodSpec) (*ApplicationDigest, error)
	GetLocalPodUUIDs() []string
	GetPodData(uuid string) (*api.Pod, error)
	Migrate(uuid string, targetNodeID string) error
	Delete(uuid string) error
	Cleanup(uuid string) error

	manageContainer(ctx context.Context, uuid string) error
}

type managerImpl struct {
	cri       cri.CRI
	kvs       KvsDriver
	messaging MessagingDriver
	// key: Pod UUID, value: sandbox ID
	localNid     string
	sandboxIdMap map[string]string
}

func NewManager(cri cri.CRI, kvs KvsDriver, messaging MessagingDriver, localNid string) Manager {
	return &managerImpl{
		cri:          cri,
		kvs:          kvs,
		messaging:    messaging,
		localNid:     localNid,
		sandboxIdMap: make(map[string]string),
	}
}

func (mgr *managerImpl) DealLocalResource(raw []byte) error {
	pod := &api.Pod{}
	err := json.Unmarshal(raw, pod)
	if err != nil {
		return err
	}

	// check deletion
	if len(pod.Meta.DeletionTimestamp) != 0 {
		if len(pod.Status.RunningNode) == 0 || mgr.getContainerStateDigest(pod) == api.ContainerStateTerminated {
			mgr.kvs.deletePod(pod.Meta.Uuid)
			return nil
		}

		return mgr.messaging.encouragePod(pod.Status.RunningNode, pod)
	}

	/*
		err = mgr.accountMgr.BindPod(&pod)
		if err != nil {
			return err
		}
		//*/

	// waiting to schedule
	if len(pod.Status.RunningNode) == 0 {
		return mgr.schedulePod(pod)
	}

	if pod.Status.RunningNode == pod.Status.TargetNode {
		stateDigest := mgr.getContainerStateDigest(pod)
		if stateDigest == api.ContainerStateTerminated || stateDigest == api.ContainerStateUnknown {
			// TODO restart pod by the restart policy
			return nil
		}
	} else {
		stateDigest := mgr.getContainerStateDigest(pod)
		if stateDigest == api.ContainerStateTerminated {
			pod.Status.RunningNode = pod.Status.TargetNode
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerStatus.ContainerID = ""
				containerStatus.Image = ""
				containerStatus.State = api.ContainerStateWaiting
			}
			return mgr.kvs.updatePod(pod)

		} else if stateDigest == api.ContainerStateUnknown {
			// TODO restart pod by the restart policy
		}
	}

	err = mgr.messaging.encouragePod(pod.Status.RunningNode, pod)
	if err != nil {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStatus.State = api.ContainerStateUnknown
		}
		return mgr.kvs.updatePod(pod)
	}

	return nil
}

func (mgr *managerImpl) Loop(ctx context.Context) error {
	return nil
}

func (mgr *managerImpl) Create(name, owner, creatorNode string, spec *api.PodSpec) (*ApplicationDigest, error) {
	pod := &api.Pod{
		Meta: &api.ObjectMeta{
			Type:        api.ResourceTypePod,
			Name:        name,
			Owner:       owner,
			CreatorNode: creatorNode,
			Uuid:        api.GeneratePodUuid(),
		},
		Spec: mgr.setDefaultPodSpec(spec),
		Status: &api.PodStatus{
			ContainerStatuses: make([]api.ContainerStatus, 0),
		},
	}

	for _ = range pod.Spec.Containers {
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
			api.ContainerStatus{
				State: api.ContainerStateWaiting,
			})
	}

	err := mgr.kvs.createPod(pod)
	// TODO: retry only if the same uuid id exists
	if err != nil {
		return nil, err
	}

	return &ApplicationDigest{
		Name:  name,
		Uuid:  pod.Meta.Uuid,
		Owner: pod.Meta.Owner,
		State: mgr.getContainerStateMessage(pod),
	}, nil
}

func (mgr *managerImpl) setDefaultPodSpec(spec *api.PodSpec) *api.PodSpec {
	if spec.Scheduler == nil {
		spec.Scheduler = &api.SchedulerSpec{
			Type: "creator",
		}
	}
	return spec
}

func (mgr *managerImpl) schedulePod(pod *api.Pod) error {
	if len(pod.Status.RunningNode) != 0 {
		return nil
	}

	switch pod.Spec.Scheduler.Type {
	case "creator":
		pod.Status.RunningNode = pod.Meta.CreatorNode
		pod.Status.TargetNode = pod.Meta.CreatorNode
		return mgr.kvs.updatePod(pod)

	default:
		return fmt.Errorf("unsupported scheduling policy:%s", pod.Spec.Scheduler.Type)
	}
}

func (mgr *managerImpl) getContainerStateDigest(pod *api.Pod) api.ContainerState {
	allTerminated := true
	hasRunning := false

	for _, containerState := range pod.Status.ContainerStatuses {
		switch containerState.State {
		case api.ContainerStateWaiting:
			allTerminated = false
		case api.ContainerStateRunning:
			allTerminated = false
			hasRunning = true
		case api.ContainerStateTerminated:
		case api.ContainerStateUnknown:
			return api.ContainerStateUnknown
		}
	}

	if allTerminated {
		return api.ContainerStateTerminated
	} else if hasRunning {
		return api.ContainerStateRunning
	} else {
		return api.ContainerStateWaiting
	}
}

func (mgr *managerImpl) getContainerStateMessage(pod *api.Pod) string {
	digest := mgr.getContainerStateDigest(pod)

	if digest == api.ContainerStateRunning {
		all := len(pod.Status.ContainerStatuses)
		running := 0

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State == api.ContainerStateRunning {
				running += 1
			}
		}

		return fmt.Sprintf("%s (%d/%d)", string(digest), running, all)
	}

	return string(digest)
}

func (mgr *managerImpl) GetLocalPodUUIDs() []string {
	uuids := make([]string, len(mgr.sandboxIdMap))
	for uuid := range mgr.sandboxIdMap {
		uuids = append(uuids, uuid)
	}
	return uuids
}

func (mgr *managerImpl) GetPodData(uuid string) (*api.Pod, error) {
	return mgr.kvs.getPod(uuid)
}

func (mgr *managerImpl) Migrate(uuid string, targetNodeID string) error {
	pod, err := mgr.kvs.getPod(uuid)
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

	return mgr.kvs.updatePod(pod)
}

func (mgr *managerImpl) Delete(uuid string) error {
	pod, err := mgr.kvs.getPod(uuid)
	if err != nil {
		return err
	}

	if len(pod.Meta.DeletionTimestamp) != 0 {
		pod.Meta.DeletionTimestamp = misc.GetTimestamp()
		return mgr.kvs.updatePod(pod)
	}

	return nil
}

func (mgr *managerImpl) Cleanup(uuid string) error {
	pod, err := mgr.kvs.getPod(uuid)
	if err != nil {
		return err
	}

	if mgr.getContainerStateDigest(pod) != api.ContainerStateUnknown {
		return fmt.Errorf("target pod of cleanup should be unknown state")
	}

	return mgr.kvs.deletePod(uuid)
}

func (mgr *managerImpl) manageContainer(ctx context.Context, uuid string) error {
	pod, err := mgr.kvs.getPod(uuid)
	if err != nil {
		return err
	}

	// force stop container if running node is not this node
	if pod.Status.RunningNode != mgr.localNid {
		sandboxId, ok := mgr.sandboxIdMap[pod.Meta.Uuid]
		if !ok {
			return nil
		}

		cl, err := mgr.cri.ListContainers(&cri.ListContainersRequest{
			Filter: &cri.ContainerFilter{
				PodSandboxId: sandboxId,
			},
		})
		if err != nil {
			return err
		}
		for _, container := range cl.Containers {
			mgr.cri.RemoveContainer(&cri.RemoveContainerRequest{
				ContainerId: container.ID,
			})
			_, err := mgr.cri.RemovePodSandbox(&cri.RemovePodSandboxRequest{
				PodSandboxId: sandboxId,
			})
			if err == nil {
				delete(mgr.sandboxIdMap, pod.Meta.Uuid)
			}
		}

		return nil
	}

	if len(pod.Meta.DeletionTimestamp) != 0 {
		// TODO: send exit signal when any container running
		// TODO: skip processing when all container exited
		// TODO: force exit all containers after timeout
	}

	sandboxId, sandboxExists := mgr.sandboxIdMap[pod.Meta.Uuid]

	// create sandbox if it isn't exist
	if !sandboxExists {
		res, err := mgr.cri.RunPodSandbox(&cri.RunPodSandboxRequest{
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

		sandboxId = res.PodSandboxId
		mgr.sandboxIdMap[pod.Meta.Uuid] = sandboxId
	}

	// start containers if they are not exist
	sandboxStatus, err := mgr.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
		PodSandboxId: sandboxId,
	})
	if err != nil {
		return err
	}

	if len(sandboxStatus.ContainersStatuses) < len(pod.Spec.Containers) {
		// check and pull image
		images, err := mgr.cri.ListImages(&cri.ListImagesRequest{})
		if err != nil {
			return err
		}

		for idx, container := range pod.Spec.Containers {
			// skip if container created
			if len(sandboxStatus.ContainersStatuses) > idx {
				continue
			}

			imageExists := false
			for _, image := range images.Images {
				if container.Image == image.Spec.Image {
					imageExists = true
					break
				}
			}
			if !imageExists {
				_, err := mgr.cri.PullImage(&cri.PullImageRequest{
					Image: cri.ImageSpec{
						Image: container.Image,
					},
				})
				if err != nil {
					return err
				}
			}

		}

		// create containers
		for idx, container := range pod.Spec.Containers {
			// skip if container created
			if len(sandboxStatus.ContainersStatuses) > idx {
				continue
			}

			envs := []cri.KeyValue{}
			for _, one := range container.Env {
				envs = append(envs, cri.KeyValue{
					Key:   one.Name,
					Value: one.Value,
				})
			}

			res, err := mgr.cri.CreateContainer(&cri.CreateContainerRequest{
				PodSandboxId: sandboxId,
				Config: cri.ContainerConfig{
					Metadata: cri.ContainerMetadata{
						Name: container.Name,
					},
					Image: cri.ImageSpec{
						Image: container.Image,
					},
					Runtime: container.Runtime,
					Args:    container.Args,
					Envs:    envs,
				},
			})

			if err != nil {
				return err
			}

			_, err = mgr.cri.StartContainer(&cri.StartContainerRequest{
				ContainerId: res.ContainerId,
			})

			if err != nil {
				return err
			}
		}

		// skip post processed when start containers
		return nil
	}

	// check container status
	hasNormalContainer := false
	hasExitedContainer := false
	for _, containerStatus := range sandboxStatus.ContainersStatuses {
		if containerStatus.State == cri.ContainerExited {
			hasExitedContainer = true
		} else {
			hasNormalContainer = true
		}
	}
	// stop all container when any container exited
	if hasNormalContainer && hasExitedContainer {
		for _, containerStatus := range sandboxStatus.ContainersStatuses {
			if containerStatus.State == cri.ContainerExited {
				continue
			}
			_, err := mgr.cri.StopContainer(&cri.StopContainerRequest{
				ContainerId: containerStatus.ID,
			})

			if err != nil {
				return err
			}
		}
	}

	// change pod phase to exit when all container exited
	if !hasNormalContainer {
		pod.Status.Phase = api.PodPhaseExited
		err = mgr.kvs.updatePod(pod)
		if err != nil {
			return err
		}
	}

	if pod.Status.Phase == api.PodPhaseMigrating {
		// TODO: start to dump all container's data if it didn't start
		// TODO: consider the state transition
	}

	if pod.Status.Phase == api.PodPhaseExited {
		// TODO: remove all of containers in the sandbox if they are exist
		// TODO: remove the sandbox if it is exist
		// TODO: remove all dump date
		// TODO: unbind pod from the account
		// TODO: remove the pod from the kvs
	}
	return nil
}
