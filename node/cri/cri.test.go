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
package cri

import (
	"time"

	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/stretchr/testify/suite"
)

type CriTest struct {
	suite.Suite
	cri CRI
}

func NewCriTest(cl crosslink.Crosslink) *CriTest {
	return &CriTest{
		cri: NewCRI(cl),
	}
}

func (ct *CriTest) AfterTest(suiteName, testName string) {
	// cleanup containers
	containersRes, err := ct.cri.ListContainers(&ListContainersRequest{})
	ct.NoError(err)
	for _, container := range containersRes.Containers {
		_, err = ct.cri.RemoveContainer(&RemoveContainerRequest{
			ContainerId: container.ID,
		})
		ct.NoError(err)
	}

	// cleanup sandboxes
	sandboxesRes, err := ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	for _, sandbox := range sandboxesRes.Items {
		_, err = ct.cri.RemovePodSandbox(&RemovePodSandboxRequest{
			PodSandboxId: sandbox.ID,
		})
		ct.NoError(err)
	}

	// cleanup images
	imagesRes, err := ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	for _, image := range imagesRes.Images {
		_, err = ct.cri.RemoveImage(&RemoveImageRequest{
			Image: image.Spec,
		})
		ct.NoError(err)
	}
}

func (ct *CriTest) TestImage() {
	// expect the listRes empty
	listRes, err := ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(listRes.Images, 0)

	// expect there to be one image after pull a image
	pullRes, err := ct.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "https://localhost:8080/test/exit.wasm",
		},
	})
	ct.NoError(err)
	test1ID := pullRes.ImageRef

	listRes, err = ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(listRes.Images, 1)
	ct.checkImage(&listRes.Images[0], "https://localhost:8080/test/exit.wasm", test1ID)

	// expect there to be two images after pull another image
	pullRes, err = ct.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "https://localhost:8080/test/sleep.wasm",
		},
	})
	ct.NoError(err)
	test2ID := pullRes.ImageRef

	listRes, err = ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(listRes.Images, 2)

	if listRes.Images[0].Spec.Image == "https://localhost:8080/test/exit.wasm" {
		ct.checkImage(&listRes.Images[0], "https://localhost:8080/test/exit.wasm", test1ID)
		ct.checkImage(&listRes.Images[1], "https://localhost:8080/test/sleep.wasm", test2ID)
	} else {
		ct.checkImage(&listRes.Images[0], "https://localhost:8080/test/sleep.wasm", test2ID)
		ct.checkImage(&listRes.Images[1], "https://localhost:8080/test/exit.wasm", test1ID)
	}
	ct.NotEqual(listRes.Images[0].ID, listRes.Images[1].ID)

	// expect there to be two images after pull the same image
	pullRes, err = ct.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "https://localhost:8080/test/exit.wasm",
		},
	})
	ct.NoError(err)
	ct.Equal(test1ID, pullRes.ImageRef)

	listRes, err = ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(listRes.Images, 2)

	// expect there to be one image after remove the image
	_, err = ct.cri.RemoveImage(&RemoveImageRequest{
		Image: ImageSpec{
			Image: "https://localhost:8080/test/exit.wasm",
		},
	})
	ct.NoError(err)

	listRes, err = ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(listRes.Images, 1)
	ct.checkImage(&listRes.Images[0], "https://localhost:8080/test/sleep.wasm", test2ID)
}

func (ct *CriTest) TestSandbox() {
	// expect that there is not sandbox
	listRes, err := ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	ct.Len(listRes.Items, 0)

	// setup image
	_, err = ct.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "https://localhost:8080/test/exit.wasm",
		},
	})
	ct.NoError(err)

	// expect there is one sandbox after run sandbox
	runRes, err := ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid1",
				Namespace: "ns1",
			},
		},
	})
	ct.NoError(err)
	sandboxId1 := runRes.PodSandboxId

	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	ct.Len(listRes.Items, 1)

	ct.Equal(listRes.Items[0].ID, sandboxId1)
	ct.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1")
	ct.Equal(listRes.Items[0].State, SandboxReady)
	ct.checkTimestampFormat(listRes.Items[0].CreatedAt)

	// expect an error when create duplicate sandbox and there is only one sandbox
	_, err = ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid2",
				Namespace: "ns1",
			},
		},
	})
	ct.Error(err)

	_, err = ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid1",
				Namespace: "ns1",
			},
		},
	})
	ct.Error(err)

	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	ct.Len(listRes.Items, 1)
	ct.Equal(listRes.Items[0].ID, sandboxId1)

	// expect there are two sandbox after run one more sandbox
	runRes, err = ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid2",
				Namespace: "ns1",
			},
		},
	})
	ct.NoError(err)
	sandboxId2 := runRes.PodSandboxId

	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	ct.Len(listRes.Items, 2)

	if listRes.Items[0].ID == sandboxId1 {
		ct.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1")
		ct.Equal(listRes.Items[1].ID, sandboxId2)
		ct.checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox2", "uid2", "ns1")
	} else {
		ct.Equal(listRes.Items[0].ID, sandboxId2)
		ct.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1")
		ct.checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox1", "uid1", "ns1")
	}

	// expect an error when call PodSandboxStatus for pod not exist
	_, err = ct.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxId: "not exist",
	})
	ct.Error(err)

	// checking response of PodSandboxStatus
	createContainerRes, err := ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandboxId1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "containerName",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/exit.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	ct.NoError(err)
	container1 := createContainerRes.ContainerId

	statusRes, err := ct.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxId: sandboxId1,
	})
	ct.NoError(err)
	ct.Equal(sandboxId1, statusRes.Status.ID)
	ct.checkSandboxMeta(&statusRes.Status.Metadata, "sandbox1", "uid1", "ns1")
	ct.Equal(statusRes.Status.State, SandboxReady)
	ct.checkTimestampFormat(statusRes.Status.CreatedAt)
	ct.Len(statusRes.ContainersStatuses, 1)
	ct.Equal(statusRes.ContainersStatuses[0].ID, container1)
	ct.Equal(statusRes.ContainersStatuses[0].Metadata.Name, "containerName")
	ct.Equal(statusRes.ContainersStatuses[0].State, ContainerCreated)
	ct.checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt)
	ct.Len(statusRes.ContainersStatuses[0].StartedAt, 0)
	ct.Len(statusRes.ContainersStatuses[0].FinishedAt, 0)
	ct.Equal(statusRes.ContainersStatuses[0].Image.Image, "https://localhost:8080/test/exit.wasm")
	ct.NotEmpty(statusRes.ContainersStatuses[0].ImageRef)
	ct.checkTimestampFormat(statusRes.Timestamp)

	// start container
	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container1,
	})
	ct.NoError(err)

	// checking status after stopping sandbox
	_, err = ct.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	ct.NoError(err)

	ct.Eventually(func() bool {
		statusRes, err = ct.cri.PodSandboxStatus(&PodSandboxStatusRequest{
			PodSandboxId: sandboxId1,
		})
		ct.NoError(err)
		return statusRes.ContainersStatuses[0].State == ContainerExited
	}, 15*time.Second, time.Second)

	ct.Equal(statusRes.Status.State, SandboxNotReady)
	ct.checkTimestampFormat(statusRes.Status.CreatedAt)
	ct.Len(statusRes.ContainersStatuses, 1)
	ct.checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt)
	ct.checkTimestampFormat(statusRes.ContainersStatuses[0].StartedAt)
	ct.checkTimestampFormat(statusRes.ContainersStatuses[0].FinishedAt)
	ct.checkTimestampFormat(statusRes.Timestamp)

	// StopPodSandbox is idempotent
	_, err = ct.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	ct.NoError(err)

	// expect there is one sandbox after remove one sandbox
	_, err = ct.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	ct.NoError(err)

	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{})
	ct.NoError(err)
	ct.Len(listRes.Items, 1)
	ct.Equal(listRes.Items[0].ID, sandboxId2)
	ct.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1")

	// RemovePodSandbox is idempotent
	_, err = ct.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	ct.NoError(err)

	// expect there is no containers
	listContainersRes, err := ct.cri.ListContainers(&ListContainersRequest{})
	ct.NoError(err)
	ct.Len(listContainersRes.Containers, 0)

	// check filter of ListPodSandbox
	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{
		Filter: &PodSandboxFilter{
			ID: sandboxId1,
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Items, 0)

	listRes, err = ct.cri.ListPodSandbox(&ListPodSandboxRequest{
		Filter: &PodSandboxFilter{
			State: &PodSandboxStateValue{
				State: SandboxNotReady,
			},
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Items, 0)

	// recreate sandbox with the same name
	_, err = ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid1",
				Namespace: "ns1",
			},
		},
	})
	ct.NoError(err)
}

func (ct *CriTest) TestContainer() {
	// expect there is no container
	listRes, err := ct.cri.ListContainers(&ListContainersRequest{})
	ct.NoError(err)
	ct.Len(listRes.Containers, 0)

	// expect there is no image
	imageListRes, err := ct.cri.ListImages(&ListImagesRequest{})
	ct.NoError(err)
	ct.Len(imageListRes.Images, 0)

	// create a pod sandbox to test
	sandboxRes, err := ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid1",
				Namespace: "ns",
			},
		},
	})
	ct.NoError(err)
	sandbox1 := sandboxRes.PodSandboxId

	sandboxRes, err = ct.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid2",
				Namespace: "ns",
			},
		},
	})
	ct.NoError(err)
	sandbox2 := sandboxRes.PodSandboxId

	// expect an error when create before pulling image
	_, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	ct.Error(err)

	// expect no error after pulling image
	for _, image := range []string{
		"https://localhost:8080/test/exit.wasm",
		"https://localhost:8080/test/sleep.wasm",
	} {
		_, err = ct.cri.PullImage(&PullImageRequest{
			Image: ImageSpec{
				Image: image,
			},
		})
		ct.NoError(err)
	}

	createRes, err := ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	ct.NoError(err)
	container1 := createRes.ContainerId

	statusRes, err := ct.cri.ContainerStatus(&ContainerStatusRequest{
		ContainerId: container1,
	})
	ct.NoError(err)
	ct.Equal(container1, statusRes.Status.ID)
	ct.Equal("container1", statusRes.Status.Metadata.Name)
	ct.Equal(ContainerCreated, statusRes.Status.State)
	ct.checkTimestampFormat(statusRes.Status.CreatedAt)
	ct.Empty(statusRes.Status.StartedAt)
	ct.Empty(statusRes.Status.FinishedAt)
	ct.Equal(0, statusRes.Status.ExitCode)
	ct.Equal("https://localhost:8080/test/sleep.wasm", statusRes.Status.Image.Image)
	ct.NotEmpty(statusRes.Status.ImageRef)

	// expect container status is running after start the container
	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container1,
	})
	ct.NoError(err)

	statusRes, err = ct.cri.ContainerStatus(&ContainerStatusRequest{
		ContainerId: container1,
	})
	ct.NoError(err)
	ct.Equal(ContainerRunning, statusRes.Status.State)
	ct.checkTimestampFormat(statusRes.Status.StartedAt)

	// expect finish program eventually
	ct.Eventually(func() bool {
		statusRes, err = ct.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container1,
		})
		ct.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	ct.checkTimestampFormat(statusRes.Status.FinishedAt)
	ct.Equal(0, statusRes.Status.ExitCode)

	// expect an error when try to create container with existing name on the same sandbox
	_, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	ct.Error(err)

	// can run container with different name from existing one on the same sandbox
	createRes, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container2",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/exit.wasm",
			},
			Runtime: []string{
				"go:1.19",
			},
			Args: []string{
				"1",
			},
		},
	})
	ct.NoError(err)
	container2 := createRes.ContainerId

	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container2,
	})
	ct.NoError(err)

	// expect finish program eventually and set exit code
	ct.Eventually(func() bool {
		statusRes, err = ct.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container2,
		})
		ct.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	ct.checkTimestampFormat(statusRes.Status.FinishedAt)
	ct.Equal(1, statusRes.Status.ExitCode)

	// can run container with the same name from existing one on the different sandbox
	createRes, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
			Args:    []string{"-1"},
		},
	})
	ct.NoError(err)
	container3 := createRes.ContainerId

	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container3,
	})
	ct.NoError(err)

	// stop container force and get error code eventually
	_, err = ct.cri.StopContainer(&StopContainerRequest{
		ContainerId: container3,
	})
	ct.NoError(err)

	ct.Eventually(func() bool {
		statusRes, err = ct.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container3,
		})
		ct.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	ct.checkTimestampFormat(statusRes.Status.FinishedAt)
	ct.Equal(137, statusRes.Status.ExitCode)

	// StopContainer is idempotent
	_, err = ct.cri.StopContainer(&StopContainerRequest{
		ContainerId: container3,
	})
	ct.NoError(err)

	// check response of list container
	listRes, err = ct.cri.ListContainers(&ListContainersRequest{})
	ct.NoError(err)
	ct.Len(listRes.Containers, 3)
	for _, container := range listRes.Containers {
		switch container.ID {
		case container1:
			ct.checkContainer(&container, container1, sandbox1, "container1", "https://localhost:8080/test/sleep.wasm", ContainerExited)
		case container2:
			ct.checkContainer(&container, container2, sandbox1, "container2", "https://localhost:8080/test/exit.wasm", ContainerExited)
		case container3:
			ct.checkContainer(&container, container3, sandbox2, "container1", "https://localhost:8080/test/sleep.wasm", ContainerExited)
		}
	}

	// check response of list container with specify id
	listRes, err = ct.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			ID: container1,
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Containers, 1)
	ct.Equal(container1, listRes.Containers[0].ID)

	// start a container with sleep infinity
	createRes, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container2",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
			Args:    []string{"-1"},
		},
	})
	ct.NoError(err)
	container4 := createRes.ContainerId

	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container4,
	})
	ct.NoError(err)

	// check response of list container with specify state
	listRes, err = ct.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			State: &ContainerStateValue{
				State: ContainerRunning,
			},
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Containers, 1)
	ct.Equal(container4, listRes.Containers[0].ID)

	// check response of list container with specify sandbox
	listRes, err = ct.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			PodSandboxId: sandbox1,
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Containers, 2)
	if listRes.Containers[0].ID == container1 {
		ct.Equal(container1, listRes.Containers[0].ID)
		ct.Equal(container2, listRes.Containers[1].ID)
	} else {
		ct.Equal(container1, listRes.Containers[1].ID)
		ct.Equal(container2, listRes.Containers[0].ID)
	}
	ct.Equal(sandbox1, listRes.Containers[0].PodSandboxId)
	ct.Equal(sandbox1, listRes.Containers[1].PodSandboxId)

	// check response of list container with specify state & sandbox
	listRes, err = ct.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			State: &ContainerStateValue{
				State: ContainerExited,
			},
			PodSandboxId: sandbox2,
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Containers, 1)
	ct.Equal(container3, listRes.Containers[0].ID)

	// check response of list container after remove a container
	_, err = ct.cri.RemoveContainer(&RemoveContainerRequest{
		ContainerId: container4,
	})
	ct.NoError(err)

	listRes, err = ct.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			PodSandboxId: sandbox2,
		},
	})
	ct.NoError(err)
	ct.Len(listRes.Containers, 1)
	ct.Equal(container3, listRes.Containers[0].ID)

	// RemoveContainer is idempotent
	_, err = ct.cri.RemoveContainer(&RemoveContainerRequest{
		ContainerId: container4,
	})
	ct.NoError(err)

	// check the container become error when specify incorrect runtime
	createRes, err = ct.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container3",
			},
			Image: ImageSpec{
				Image: "https://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"incorrect"},
		},
	})
	ct.NoError(err)
	container5 := createRes.ContainerId

	_, err = ct.cri.StartContainer(&StartContainerRequest{
		ContainerId: container5,
	})
	ct.NoError(err)

	ct.Eventually(func() bool {
		statusRes, err = ct.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container5,
		})
		ct.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	ct.Equal(-1, statusRes.Status.ExitCode)
}

func (ct *CriTest) checkImage(image *Image, url string, id string) {
	ct.NotEmpty(image.ID)
	ct.Equal(id, image.ID)
	ct.Equal(url, image.Spec.Image)
}

// check timestamp format (ISO8601/RFC3339)
func (ct *CriTest) checkTimestampFormat(timestamp string) {
	_, err := time.Parse(time.RFC3339, timestamp)
	ct.NoError(err)
}

func (ct *CriTest) checkSandboxMeta(meta *PodSandboxMetadata, name, uid, namespace string) {
	ct.Equal(name, meta.Name)
	ct.NotEmpty(meta.UID)
	ct.Equal(uid, meta.UID)
	ct.Equal(namespace, meta.Namespace)
}

func (ct *CriTest) checkContainer(container *Container, id, sandbox, name, url string, state ContainerState) {
	ct.Equal(id, container.ID)
	ct.Equal(sandbox, container.PodSandboxId)
	ct.Equal(name, container.Metadata.Name)
	ct.Equal(url, container.Image.Image)
	ct.NotEmpty(container.ImageRef)
	ct.Equal(state, container.State)
	ct.checkTimestampFormat(container.CreatedAt)
}
