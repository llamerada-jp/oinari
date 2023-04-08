package cri

import (
	"time"

	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	cri CRI
}

func NewTestSuite(cl crosslink.Crosslink) *TestSuite {
	return &TestSuite{
		cri: NewCRI(cl),
	}
}

func (suite *TestSuite) AfterTest(suiteName, testName string) {
	// cleanup containers
	containersRes, err := suite.cri.ListContainers(&ListContainersRequest{})
	suite.NoError(err)
	for _, container := range containersRes.Containers {
		_, err = suite.cri.RemoveContainer(&RemoveContainerRequest{
			ContainerId: container.ID,
		})
		suite.NoError(err)
	}

	// cleanup sandboxes
	sandboxesRes, err := suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	for _, sandbox := range sandboxesRes.Items {
		_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
			PodSandboxId: sandbox.ID,
		})
		suite.NoError(err)
	}

	// cleanup images
	imagesRes, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	for _, image := range imagesRes.Images {
		_, err = suite.cri.RemoveImage(&RemoveImageRequest{
			Image: image.Spec,
		})
		suite.NoError(err)
	}
}

func (suite *TestSuite) TestImage() {
	// expect the listRes empty
	listRes, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 0)

	// expect there to be one image after pull a image
	pullRes, err := suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test/exit.wasm",
		},
	})
	suite.NoError(err)
	test1ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.checkImage(&listRes.Images[0], "http://localhost:8080/test/exit.wasm", test1ID)

	// expect there to be two images after pull another image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test/sleep.wasm",
		},
	})
	suite.NoError(err)
	test2ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 2)

	if listRes.Images[0].Spec.Image == "http://localhost:8080/test/exit.wasm" {
		suite.checkImage(&listRes.Images[0], "http://localhost:8080/test/exit.wasm", test1ID)
		suite.checkImage(&listRes.Images[1], "http://localhost:8080/test/sleep.wasm", test2ID)
	} else {
		suite.checkImage(&listRes.Images[0], "http://localhost:8080/test/sleep.wasm", test2ID)
		suite.checkImage(&listRes.Images[1], "http://localhost:8080/test/exit.wasm", test1ID)
	}
	suite.NotEqual(listRes.Images[0].ID, listRes.Images[1].ID)

	// expect there to be two images after pull the same image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test/exit.wasm",
		},
	})
	suite.NoError(err)
	suite.Equal(test1ID, pullRes.ImageRef)

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 2)

	// expect there to be one image after remove the image
	_, err = suite.cri.RemoveImage(&RemoveImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test/exit.wasm",
		},
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.checkImage(&listRes.Images[0], "http://localhost:8080/test/sleep.wasm", test2ID)
}

func (suite *TestSuite) TestSandbox() {
	// expect that there is not sandbox
	listRes, err := suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 0)

	// setup image
	_, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test/exit.wasm",
		},
	})
	suite.NoError(err)

	// expect there is one sandbox after run sandbox
	runRes, err := suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid1",
				Namespace: "ns1",
			},
		},
	})
	suite.NoError(err)
	sandboxId1 := runRes.PodSandboxId

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 1)

	suite.Equal(listRes.Items[0].ID, sandboxId1)
	suite.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1")
	suite.Equal(listRes.Items[0].State, SandboxReady)
	suite.checkTimestampFormat(listRes.Items[0].CreatedAt)

	// expect an error when create duplicate sandbox and there is only one sandbox
	_, err = suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid2",
				Namespace: "ns1",
			},
		},
	})
	suite.Error(err)

	_, err = suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid1",
				Namespace: "ns1",
			},
		},
	})
	suite.Error(err)

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 1)
	suite.Equal(listRes.Items[0].ID, sandboxId1)

	// expect there are two sandbox after run one more sandbox
	runRes, err = suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid2",
				Namespace: "ns1",
			},
		},
	})
	suite.NoError(err)
	sandboxId2 := runRes.PodSandboxId

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 2)

	if listRes.Items[0].ID == sandboxId1 {
		suite.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1")
		suite.Equal(listRes.Items[1].ID, sandboxId2)
		suite.checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox2", "uid2", "ns1")
	} else {
		suite.Equal(listRes.Items[0].ID, sandboxId2)
		suite.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1")
		suite.checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox1", "uid1", "ns1")
	}

	// expect an error when call PodSandboxStatus for pod not exist
	_, err = suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxId: "not exist",
	})
	suite.Error(err)

	// checking response of PodSandboxStatus
	createContainerRes, err := suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandboxId1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "containerName",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/exit.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	suite.NoError(err)
	container1 := createContainerRes.ContainerId

	statusRes, err := suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxId: sandboxId1,
	})
	suite.NoError(err)
	suite.Equal(sandboxId1, statusRes.Status.ID)
	suite.checkSandboxMeta(&statusRes.Status.Metadata, "sandbox1", "uid1", "ns1")
	suite.Equal(statusRes.Status.State, SandboxReady)
	suite.checkTimestampFormat(statusRes.Status.CreatedAt)
	suite.Len(statusRes.ContainersStatuses, 1)
	suite.Equal(statusRes.ContainersStatuses[0].ID, container1)
	suite.Equal(statusRes.ContainersStatuses[0].Metadata.Name, "containerName")
	suite.Equal(statusRes.ContainersStatuses[0].State, ContainerCreated)
	suite.checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt)
	suite.Len(statusRes.ContainersStatuses[0].StartedAt, 0)
	suite.Len(statusRes.ContainersStatuses[0].FinishedAt, 0)
	suite.Equal(statusRes.ContainersStatuses[0].Image.Image, "http://localhost:8080/test/exit.wasm")
	suite.NotEmpty(statusRes.ContainersStatuses[0].ImageRef)
	suite.checkTimestampFormat(statusRes.Timestamp)

	// start container
	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container1,
	})
	suite.NoError(err)

	// checking status after stopping sandbox
	_, err = suite.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	suite.NoError(err)

	suite.Eventually(func() bool {
		statusRes, err = suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
			PodSandboxId: sandboxId1,
		})
		suite.NoError(err)
		return statusRes.ContainersStatuses[0].State == ContainerExited
	}, 15*time.Second, time.Second)

	suite.Equal(statusRes.Status.State, SandboxNotReady)
	suite.checkTimestampFormat(statusRes.Status.CreatedAt)
	suite.Len(statusRes.ContainersStatuses, 1)
	suite.checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt)
	suite.checkTimestampFormat(statusRes.ContainersStatuses[0].StartedAt)
	suite.checkTimestampFormat(statusRes.ContainersStatuses[0].FinishedAt)
	suite.checkTimestampFormat(statusRes.Timestamp)

	// StopPodSandbox is idempotent
	_, err = suite.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	suite.NoError(err)

	// expect there is one sandbox after remove one sandbox
	_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 1)
	suite.Equal(listRes.Items[0].ID, sandboxId2)
	suite.checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1")

	// RemovePodSandbox is idempotent
	_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxId: sandboxId1,
	})
	suite.NoError(err)

	// expect there is no containers
	listContainersRes, err := suite.cri.ListContainers(&ListContainersRequest{})
	suite.NoError(err)
	suite.Len(listContainersRes.Containers, 0)

	// check filter of ListPodSandbox
	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{
		Filter: &PodSandboxFilter{
			ID: sandboxId1,
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Items, 0)

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{
		Filter: &PodSandboxFilter{
			State: &PodSandboxStateValue{
				State: SandboxNotReady,
			},
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Items, 0)
}

func (suite *TestSuite) TestContainer() {
	// expect there is no container
	listRes, err := suite.cri.ListContainers(&ListContainersRequest{})
	suite.NoError(err)
	suite.Len(listRes.Containers, 0)

	// expect there is no image
	imageListRes, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(imageListRes.Images, 0)

	// create a pod sandbox to test
	sandboxRes, err := suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox1",
				UID:       "uid1",
				Namespace: "ns",
			},
		},
	})
	suite.NoError(err)
	sandbox1 := sandboxRes.PodSandboxId

	sandboxRes, err = suite.cri.RunPodSandbox(&RunPodSandboxRequest{
		Config: PodSandboxConfig{
			Metadata: PodSandboxMetadata{
				Name:      "sandbox2",
				UID:       "uid2",
				Namespace: "ns",
			},
		},
	})
	suite.NoError(err)
	sandbox2 := sandboxRes.PodSandboxId

	// expect an error when create before pulling image
	_, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	suite.Error(err)

	// expect no error after pulling image
	for _, image := range []string{
		"http://localhost:8080/test/exit.wasm",
		"http://localhost:8080/test/sleep.wasm",
	} {
		_, err = suite.cri.PullImage(&PullImageRequest{
			Image: ImageSpec{
				Image: image,
			},
		})
		suite.NoError(err)
	}

	createRes, err := suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	suite.NoError(err)
	container1 := createRes.ContainerId

	statusRes, err := suite.cri.ContainerStatus(&ContainerStatusRequest{
		ContainerId: container1,
	})
	suite.NoError(err)
	suite.Equal(container1, statusRes.Status.ID)
	suite.Equal("container1", statusRes.Status.Metadata.Name)
	suite.Equal(ContainerCreated, statusRes.Status.State)
	suite.checkTimestampFormat(statusRes.Status.CreatedAt)
	suite.Empty(statusRes.Status.StartedAt)
	suite.Empty(statusRes.Status.FinishedAt)
	suite.Equal(0, statusRes.Status.ExitCode)
	suite.Equal("http://localhost:8080/test/sleep.wasm", statusRes.Status.Image.Image)
	suite.NotEmpty(statusRes.Status.ImageRef)

	// expect container status is running after start the container
	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container1,
	})
	suite.NoError(err)

	statusRes, err = suite.cri.ContainerStatus(&ContainerStatusRequest{
		ContainerId: container1,
	})
	suite.NoError(err)
	suite.Equal(ContainerRunning, statusRes.Status.State)
	suite.checkTimestampFormat(statusRes.Status.StartedAt)

	// expect finish program eventually
	suite.Eventually(func() bool {
		statusRes, err = suite.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container1,
		})
		suite.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	suite.checkTimestampFormat(statusRes.Status.FinishedAt)
	suite.Equal(0, statusRes.Status.ExitCode)

	// expect an error when try to create container with existing name on the same sandbox
	_, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
		},
	})
	suite.Error(err)

	// can run container with different name from existing one on the same sandbox
	createRes, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container2",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/exit.wasm",
			},
			Runtime: []string{
				"go:1.19",
			},
			Args: []string{
				"1",
			},
		},
	})
	suite.NoError(err)
	container2 := createRes.ContainerId

	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container2,
	})
	suite.NoError(err)

	// expect finish program eventually and set exit code
	suite.Eventually(func() bool {
		statusRes, err = suite.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container2,
		})
		suite.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	suite.checkTimestampFormat(statusRes.Status.FinishedAt)
	suite.Equal(1, statusRes.Status.ExitCode)

	// can run container with the same name from existing one on the different sandbox
	createRes, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container1",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
			Args:    []string{"-1"},
		},
	})
	suite.NoError(err)
	container3 := createRes.ContainerId

	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container3,
	})
	suite.NoError(err)

	// stop container force and get error code eventually
	_, err = suite.cri.StopContainer(&StopContainerRequest{
		ContainerId: container3,
	})
	suite.NoError(err)

	suite.Eventually(func() bool {
		statusRes, err = suite.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container3,
		})
		suite.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	suite.checkTimestampFormat(statusRes.Status.FinishedAt)
	suite.Equal(137, statusRes.Status.ExitCode)

	// StopContainer is idempotent
	_, err = suite.cri.StopContainer(&StopContainerRequest{
		ContainerId: container3,
	})
	suite.NoError(err)

	// check response of list container
	listRes, err = suite.cri.ListContainers(&ListContainersRequest{})
	suite.NoError(err)
	suite.Len(listRes.Containers, 3)
	for _, container := range listRes.Containers {
		switch container.ID {
		case container1:
			suite.checkContainer(&container, container1, sandbox1, "container1", "http://localhost:8080/test/sleep.wasm", ContainerExited)
		case container2:
			suite.checkContainer(&container, container2, sandbox1, "container2", "http://localhost:8080/test/exit.wasm", ContainerExited)
		case container3:
			suite.checkContainer(&container, container3, sandbox2, "container1", "http://localhost:8080/test/sleep.wasm", ContainerExited)
		}
	}

	// check response of list container with specify id
	listRes, err = suite.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			ID: container1,
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Containers, 1)
	suite.Equal(container1, listRes.Containers[0].ID)

	// start a container with sleep infinity
	createRes, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container2",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"go:1.19"},
			Args:    []string{"-1"},
		},
	})
	suite.NoError(err)
	container4 := createRes.ContainerId

	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container4,
	})
	suite.NoError(err)

	// check response of list container with specify state
	listRes, err = suite.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			State: &ContainerStateValue{
				State: ContainerRunning,
			},
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Containers, 1)
	suite.Equal(container4, listRes.Containers[0].ID)

	// check response of list container with specify sandbox
	listRes, err = suite.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			PodSandboxId: sandbox1,
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Containers, 2)
	if listRes.Containers[0].ID == container1 {
		suite.Equal(container1, listRes.Containers[0].ID)
		suite.Equal(container2, listRes.Containers[1].ID)
	} else {
		suite.Equal(container1, listRes.Containers[1].ID)
		suite.Equal(container2, listRes.Containers[0].ID)
	}
	suite.Equal(sandbox1, listRes.Containers[0].PodSandboxId)
	suite.Equal(sandbox1, listRes.Containers[1].PodSandboxId)

	// check response of list container with specify state & sandbox
	listRes, err = suite.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			State: &ContainerStateValue{
				State: ContainerExited,
			},
			PodSandboxId: sandbox2,
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Containers, 1)
	suite.Equal(container3, listRes.Containers[0].ID)

	// check response of list container after remove a container
	_, err = suite.cri.RemoveContainer(&RemoveContainerRequest{
		ContainerId: container4,
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListContainers(&ListContainersRequest{
		Filter: &ContainerFilter{
			PodSandboxId: sandbox2,
		},
	})
	suite.NoError(err)
	suite.Len(listRes.Containers, 1)
	suite.Equal(container3, listRes.Containers[0].ID)

	// RemoveContainer is idempotent
	_, err = suite.cri.RemoveContainer(&RemoveContainerRequest{
		ContainerId: container4,
	})
	suite.NoError(err)

	// check the container become error when specify incorrect runtime
	createRes, err = suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxId: sandbox2,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "container3",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test/sleep.wasm",
			},
			Runtime: []string{"incorrect"},
		},
	})
	suite.NoError(err)
	container5 := createRes.ContainerId

	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerId: container5,
	})
	suite.NoError(err)

	suite.Eventually(func() bool {
		statusRes, err = suite.cri.ContainerStatus(&ContainerStatusRequest{
			ContainerId: container5,
		})
		suite.NoError(err)
		return statusRes.Status.State == ContainerExited
	}, 15*time.Second, time.Second)
	suite.Equal(-1, statusRes.Status.ExitCode)
}

func (suite *TestSuite) checkImage(image *Image, url string, id string) {
	suite.NotEmpty(image.ID)
	suite.Equal(id, image.ID)
	suite.Equal(url, image.Spec.Image)
}

// check timestamp format (ISO8601/RFC3339)
func (suite *TestSuite) checkTimestampFormat(timestamp string) {
	_, err := time.Parse(time.RFC3339, timestamp)
	suite.NoError(err)
}

func (suite *TestSuite) checkSandboxMeta(meta *PodSandboxMetadata, name, uid, namespace string) {
	suite.Equal(name, meta.Name)
	suite.NotEmpty(meta.UID)
	suite.Equal(uid, meta.UID)
	suite.Equal(namespace, meta.Namespace)
}

func (suite *TestSuite) checkContainer(container *Container, id, sandbox, name, url string, state ContainerState) {
	suite.Equal(id, container.ID)
	suite.Equal(sandbox, container.PodSandboxId)
	suite.Equal(name, container.Metadata.Name)
	suite.Equal(url, container.Image.Image)
	suite.NotEmpty(container.ImageRef)
	suite.Equal(state, container.State)
	suite.checkTimestampFormat(container.CreatedAt)
}
