package cri

import (
	"time"

	"github.com/llamerada-jp/oinari/agent/crosslink"
	"github.com/stretchr/testify/suite"
)

type CriSuite struct {
	suite.Suite
	cri CRI
}

func (suite *CriSuite) SetupSuite() {
	rootMpx := crosslink.NewMultiPlexer()
	cl := crosslink.NewCrosslink("crosslink", rootMpx)

	suite.cri = NewCRI(cl)
}

func (suite *CriSuite) TestImage() {
	// expect the listRes empty
	listRes, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Equal(0, len(listRes.Images))

	// expect there to be one image after pull a image
	pullRes, err := suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test1.container.json",
		},
	})
	suite.NoError(err)
	test1ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test1.container.json", "go1.19", test1ID))

	// expect there to be two images after pull another image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test2.container.json",
		},
	})
	suite.NoError(err)
	test2ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 2)

	if listRes.Images[0].Spec.Image == "http://localhost:8080/test1.container.json" {
		suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test1.container.json", "go1.19", test1ID))
		suite.True(checkImage(&listRes.Images[1], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
	} else {
		suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
		suite.True(checkImage(&listRes.Images[1], "http://localhost:8080/test1.container.json", "go1.19", test1ID))
	}
	suite.NotEqual(listRes.Images[0].ID, listRes.Images[1].ID)

	// expect there to be two images after pull the same image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test1.container.json",
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
			Image: "http://localhost:8080/test1.container.json",
		},
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
}

func (suite *CriSuite) TestSandbox() {
	// expect that there is not sandbox
	listRes, err := suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 0)

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
	sandboxId1 := runRes.PodSandboxID

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 1)

	suite.Equal(listRes.Items[0].ID, sandboxId1)
	suite.True(checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1"))
	suite.Equal(listRes.Items[0].State, SandboxReady)
	suite.True(checkTimestampFormat(listRes.Items[0].CreatedAt))

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
	sandboxId2 := runRes.PodSandboxID

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 2)

	if listRes.Items[0].ID == sandboxId1 {
		suite.True(checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox1", "uid1", "ns1"))
		suite.Equal(listRes.Items[1].ID, sandboxId2)
		suite.True(checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox2", "uid2", "ns1"))
	} else {
		suite.Equal(listRes.Items[0].ID, sandboxId2)
		suite.True(checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1"))
		suite.True(checkSandboxMeta(&listRes.Items[1].Metadata, "sandbox1", "uid1", "ns1"))
	}

	// expect an error when call PodSandboxStatus for pod not exist
	_, err = suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxID: "not exist",
	})
	suite.Error(err)

	// checking response of PodSandboxStatus
	createContainerRes, err := suite.cri.CreateContainer(&CreateContainerRequest{
		PodSandboxID: sandboxId1,
		Config: ContainerConfig{
			Metadata: ContainerMetadata{
				Name: "containerName",
			},
			Image: ImageSpec{
				Image: "http://localhost:8080/test1.container.json",
			},
		},
	})
	suite.NoError(err)
	container1 := createContainerRes.ContainerID

	statusRes, err := suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)
	suite.Equal(statusRes.Status.ID, sandboxId1)
	suite.True(checkSandboxMeta(&statusRes.Status.Metadata, "sandbox1", "uid1", "ns1"))
	suite.Equal(statusRes.Status.State, SandboxReady)
	suite.True(checkTimestampFormat(statusRes.Status.CreatedAt))
	suite.Len(statusRes.ContainersStatuses, 1)
	suite.Equal(statusRes.ContainersStatuses[0].ID, container1)
	suite.Equal(statusRes.ContainersStatuses[0].Metadata.Name, "containerName")
	suite.Equal(statusRes.ContainersStatuses[0].State, ContainerCreated)
	suite.True(checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt))
	suite.Len(statusRes.ContainersStatuses[0].StartedAt, 0)
	suite.Len(statusRes.ContainersStatuses[0].FinishedAt, 0)
	suite.Equal(statusRes.ContainersStatuses[0].Image.Image, "http://localhost:8080/test1.container.json")
	suite.NotEmpty(statusRes.ContainersStatuses[0].ImageRef)
	suite.True(checkTimestampFormat(statusRes.Timestamp))

	// start container
	_, err = suite.cri.StartContainer(&StartContainerRequest{
		ContainerID: container1,
	})
	suite.NoError(err)

	// checking status after stopping sandbox
	_, err = suite.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)

	statusRes, err = suite.cri.PodSandboxStatus(&PodSandboxStatusRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)
	suite.Equal(statusRes.Status.State, SandboxNotReady)
	suite.True(checkTimestampFormat(statusRes.Status.CreatedAt))
	suite.Len(statusRes.ContainersStatuses, 1)
	suite.Equal(statusRes.ContainersStatuses[0].State, ContainerExited)
	suite.True(checkTimestampFormat(statusRes.ContainersStatuses[0].CreatedAt))
	suite.True(checkTimestampFormat(statusRes.ContainersStatuses[0].StartedAt))
	suite.True(checkTimestampFormat(statusRes.ContainersStatuses[0].FinishedAt))
	suite.True(checkTimestampFormat(statusRes.Timestamp))

	// StopPodSandbox is idempotent
	_, err = suite.cri.StopPodSandbox(&StopPodSandboxRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)

	// expect there is one sandbox after remove one sandbox
	_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListPodSandbox(&ListPodSandboxRequest{})
	suite.NoError(err)
	suite.Len(listRes.Items, 1)
	suite.Equal(listRes.Items[0].ID, sandboxId2)
	suite.True(checkSandboxMeta(&listRes.Items[0].Metadata, "sandbox2", "uid2", "ns1"))

	// RemovePodSandbox is idempotent
	_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxID: sandboxId1,
	})
	suite.NoError(err)

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

	// cleanup
	_, err = suite.cri.RemovePodSandbox(&RemovePodSandboxRequest{
		PodSandboxID: sandboxId2,
	})
	suite.NoError(err)
}

func checkImage(image *Image, url, runtime, id string) bool {
	if len(image.ID) == 0 || image.ID != id ||
		len(image.Spec.Image) == 0 || image.Spec.Image != url ||
		len(image.Runtime) == 0 || image.Runtime != runtime {
		return false
	}

	return true
}

// check timestamp format (ISO8601/RFC3339)
func checkTimestampFormat(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)
	return err == nil
}

func checkSandboxMeta(meta *PodSandboxMetadata, name, uid, namespace string) bool {
	if len(meta.Name) == 0 || meta.Name != name ||
		len(meta.UID) == 0 || meta.UID != uid ||
		len(meta.Namespace) == 0 || meta.Namespace != namespace {
		return false
	}

	return true
}
