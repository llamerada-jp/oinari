package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/llamerada-jp/oinari/agent/cri"
	"github.com/llamerada-jp/oinari/agent/resource/account"
	"github.com/llamerada-jp/oinari/agent/resource/node"
	"github.com/llamerada-jp/oinari/api"
)

type Manager interface {
	DealLocalResource(raw []byte) error
	Loop(ctx context.Context) error

	run(name string, spec *api.PodSpec) (*applicationDigest, error)
	list() ([]applicationDigest, error)
	terminate(uuid string) error
	encouragePod(ctx context.Context, pod *api.Pod) error
}

type managerImpl struct {
	cri        cri.CRI
	kvs        KvsDriver
	messaging  MessagingDriver
	accountMgr account.Manager
	nodeMgr    node.Manager
	// key: Pod UUID, value: sandbox ID
	sandboxIDMap map[string]string
}

func NewManager(cri cri.CRI, kvs KvsDriver, messaging MessagingDriver, accountMgr account.Manager, nodeMgr node.Manager) Manager {
	return &managerImpl{
		cri:          cri,
		kvs:          kvs,
		messaging:    messaging,
		accountMgr:   accountMgr,
		nodeMgr:      nodeMgr,
		sandboxIDMap: make(map[string]string),
	}
}

func (mgr *managerImpl) DealLocalResource(raw []byte) error {
	var pod api.Pod
	err := json.Unmarshal(raw, &pod)
	if err != nil {
		return err
	}

	switch pod.Status.Phase {

	case api.PodPhasePending:
		err = mgr.accountMgr.BindPod(&pod)
		if err != nil {
			return err
		}

		err = mgr.schedulePod(&pod)
		if err != nil {
			return err
		}

	case api.PodPhaseRunning, api.PodPhaseMigrating:
		err = mgr.accountMgr.BindPod(&pod)
		if err != nil {
			return err
		}

		err = mgr.messaging.encouragePod(pod.Status.RunningNode, &pod)
		if err != nil {
			return err
		}

	case api.PodPhaseExited:
		err = mgr.messaging.encouragePod(pod.Status.RunningNode, &pod)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *managerImpl) Loop(ctx context.Context) error {
	return nil
}

func (mgr *managerImpl) run(name string, spec *api.PodSpec) (*applicationDigest, error) {
	pod := &api.Pod{
		Meta: &api.ObjectMeta{
			Type:        api.ResourceTypePod,
			Name:        name,
			Owner:       mgr.accountMgr.GetAccountName(),
			CreatorNode: mgr.nodeMgr.GetNid(),
			Uuid:        uuid.Must(uuid.NewRandom()).String(),
		},
		Spec: mgr.setDefaultPodSpec(spec),
		Status: &api.PodStatus{
			Phase: api.PodPhasePending,
		},
	}

	err := mgr.kvs.createPod(pod)
	// TODO: retry only if the same uuid id exists
	if err != nil {
		return nil, err
	}

	return &applicationDigest{
		Name:  name,
		Uuid:  pod.Meta.Uuid,
		Owner: pod.Meta.Owner,
	}, nil
}

func (mgr *managerImpl) list() ([]applicationDigest, error) {
	log.Fatal("TODO")
	return nil, fmt.Errorf("TODO")
}

func (mgr *managerImpl) terminate(uuid string) error {
	log.Fatal("TODO")
	return fmt.Errorf("TODO")
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
	if pod.Status.RunningNode != "" {
		return nil
	}

	switch pod.Spec.Scheduler.Type {
	case "creator":
		err := mgr.messaging.encouragePod(pod.Meta.CreatorNode, pod)
		return err

	default:
		return fmt.Errorf("unsupported scheduling policy:%s", pod.Spec.Scheduler.Type)
	}
}

func (mgr *managerImpl) encouragePod(ctx context.Context, pod *api.Pod) error {
	if pod.Status.Phase == api.PodPhasePending {
		// change pod phase to `Running` and RunningNode to this node
		pod.Status.RunningNode = mgr.nodeMgr.GetNid()

		if pod.Meta.DeletionTimestamp != "" {
			pod.Status.Phase = api.PodPhaseExited
		} else {
			pod.Status.Phase = api.PodPhaseRunning
		}

		err := mgr.kvs.updatePod(pod)
		if err != nil {
			return err
		}
	}

	sandboxID, sandboxExists := mgr.sandboxIDMap[pod.Meta.Uuid]

	if pod.Meta.DeletionTimestamp != "" {
		// TODO: send exit signal when any container running
		// TODO: skip processing when all container exited
		// TODO: force exit all containers after timeout
	}

	if pod.Status.Phase == api.PodPhaseRunning {
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

			sandboxID = res.PodSandboxID
			mgr.sandboxIDMap[pod.Meta.Uuid] = sandboxID
		}

		// start containers if they are not exist
		sandboxStatus, err := mgr.cri.PodSandboxStatus(&cri.PodSandboxStatusRequest{
			PodSandboxID: sandboxID,
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

				res, err := mgr.cri.CreateContainer(&cri.CreateContainerRequest{
					PodSandboxID: sandboxID,
					Config: cri.ContainerConfig{
						Metadata: cri.ContainerMetadata{
							Name: container.Name,
						},
						Image: cri.ImageSpec{
							Image: container.Image,
						},
					},
				})

				if err != nil {
					return err
				}

				_, err = mgr.cri.StartContainer(&cri.StartContainerRequest{
					ContainerID: res.ContainerID,
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
					ContainerID: containerStatus.ID,
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
