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
package controller

import (
	"time"

	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/kvs"
	"github.com/llamerada-jp/oinari/node/misc"
	"github.com/llamerada-jp/oinari/node/mock"
	"github.com/stretchr/testify/suite"
)

const (
	TEST_NID = ""
)

type podControllerTest struct {
	suite.Suite
	col    *mock.Colonio
	podKvs kvs.PodKvs
	mdMock *mock.MessagingDriver
	impl   *podControllerImpl
}

func NewPodControllerTest() suite.TestingSuite {
	colMock := mock.NewColonioMock()
	podKvs := kvs.NewPodKvs(colMock)
	mdMock := mock.NewMessagingDriverMock()

	return &podControllerTest{
		col:    colMock,
		podKvs: podKvs,
		mdMock: mdMock,
		impl: &podControllerImpl{
			podKvs:    podKvs,
			messaging: mdMock,
			localNid:  TEST_NID,
		},
	}
}

func (test *podControllerTest) TestDealLocalResource() {
	// require deletion if pod is not valid
	deleteFlg, err := test.impl.DealLocalResource([]byte(""))
	test.Error(err)
	test.True(deleteFlg)

	// create an pod
	nodeID1 := "012345678901234567890123456789ab"
	nodeID2 := "012345678901234567890123456789ac"
	digest, err := test.impl.Create("test-pod", "owner", nodeID1, &api.PodSpec{
		Containers: []api.ContainerSpec{
			{
				Name:          "test",
				Image:         "http://localhost/dummy.wasm",
				Runtime:       []string{"go:1.20"},
				RestartPolicy: api.RestartPolicyAlways,
			},
		},
	})
	test.NoError(err)

	// schedule node for new pod
	raw := test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.False(deleteFlg)

	pod, err := test.podKvs.Get(digest.Uuid)
	test.NoError(err)
	test.Equal(nodeID1, pod.Spec.TargetNode)
	test.Equal(nodeID1, pod.Status.RunningNode)

	test.Len(test.mdMock.Records, 0)

	raw = test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.False(deleteFlg)
	// data is not changed
	test.Equal(raw, test.getRaw(digest.Uuid))
	// rpc will be called for pod running node
	test.Len(test.mdMock.Records, 1)
	record := test.mdMock.Records[0]
	test.Equal(nodeID1, record.DestNodeID)
	test.Equal(digest.Uuid, record.ReconcileContainer.PodUuid)

	// when migrate the pod, wait to terminate containers, and will reset container status
	err = test.impl.Migrate(digest.Uuid, nodeID2)
	test.NoError(err)
	raw = test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.False(deleteFlg)
	test.Len(test.mdMock.Records, 2)
	test.Equal(raw, test.getRaw(digest.Uuid)) // waiting

	pod, err = test.podKvs.Get(digest.Uuid)
	test.NoError(err)
	test.Equal(nodeID2, pod.Spec.TargetNode)
	test.Equal(nodeID1, pod.Status.RunningNode)
	pod.Status.ContainerStatuses[0] = api.ContainerStatus{
		ContainerID: "test",
		Image:       "http://localhost/dummy.wasm",
		State: api.ContainerState{
			Running: &api.ContainerStateRunning{
				StartedAt: misc.GetTimestamp(),
			},
			Terminated: &api.ContainerStateTerminated{
				FinishedAt: misc.GetTimestamp(),
				ExitCode:   0,
			},
		},
	}
	err = test.podKvs.Update(pod) // set container terminated
	test.NoError(err)

	raw = test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.False(deleteFlg)
	test.Len(test.mdMock.Records, 2)
	pod, err = test.podKvs.Get(digest.Uuid)
	test.NoError(err)
	test.Equal(nodeID2, pod.Status.RunningNode)
	test.Equal("", pod.Status.ContainerStatuses[0].ContainerID)
	test.Equal("", pod.Status.ContainerStatuses[0].Image)
	test.Nil(pod.Status.ContainerStatuses[0].State.Running)
	test.Nil(pod.Status.ContainerStatuses[0].State.Terminated)
	test.Nil(pod.Status.ContainerStatuses[0].State.Unknown)

	// tests for pod with deletion timestamp
	digest, err = test.impl.Create("test-pod", "owner", nodeID1, &api.PodSpec{
		Containers: []api.ContainerSpec{
			{
				Name:          "test",
				Image:         "http://localhost/dummy.wasm",
				Runtime:       []string{"go:1.20"},
				RestartPolicy: api.RestartPolicyAlways,
			},
		},
	})
	test.NoError(err)

	pod, err = test.podKvs.Get(digest.Uuid)
	test.NoError(err)
	pod.Meta.DeletionTimestamp = misc.GetTimestamp()
	err = test.podKvs.Update(pod) // delete the pod
	test.NoError(err)
	// require deletion if deletion timestamp is set and container does not start
	raw = test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.True(deleteFlg)

	pod, err = test.podKvs.Get(digest.Uuid)
	test.NoError(err)
	pod.Status.RunningNode = nodeID1
	pod.Status.ContainerStatuses[0] = api.ContainerStatus{
		ContainerID: "test",
		Image:       "http://localhost/dummy.wasm",
		State: api.ContainerState{
			Running: &api.ContainerStateRunning{
				StartedAt: misc.GetTimestamp(),
			},
		},
	}
	err = test.podKvs.Update(pod) // set container running
	test.NoError(err)
	raw = test.getRaw(digest.Uuid)
	// not require deletion if deletion timestamp is set but container is running
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.False(deleteFlg)

	// require deletion if deletion timestamp is set and container terminated
	pod.Status.ContainerStatuses[0].State.Terminated = &api.ContainerStateTerminated{
		FinishedAt: misc.GetTimestamp(),
		ExitCode:   0,
	}
	err = test.podKvs.Update(pod) // set container running
	test.NoError(err)
	raw = test.getRaw(digest.Uuid)
	deleteFlg, err = test.impl.DealLocalResource(raw)
	test.NoError(err)
	test.True(deleteFlg)
}

func (test *podControllerTest) getRaw(podUuid string) []byte {
	key := string(api.ResourceTypePod) + "/" + podUuid

	val, err := test.col.KvsGet(key)

	test.NoError(err)
	raw, err := val.GetBinary()
	test.NoError(err)

	return raw
}

func (test *podControllerTest) TestPodLifecycle() {
	// create a pod
	validPodSpec := &api.PodSpec{
		Containers: []api.ContainerSpec{
			{
				Name:          "test",
				Image:         "http://localhost/dummy.wasm",
				Runtime:       []string{"go:1.20"},
				RestartPolicy: api.RestartPolicyAlways,
			},
		},
	}

	nodeID1 := "012345678901234567890123456789ab"
	nodeID2 := "012345678901234567890123456789ac"
	digest1, err := test.impl.Create("test-pod", "test-owner", nodeID1, validPodSpec)
	test.NoError(err)
	test.Equal("test-pod", digest1.Name)
	test.NotEmpty(digest1.Uuid)
	test.Equal("test-owner", digest1.Owner)
	test.NotEmpty(digest1.State)

	// create a pod with the same name
	digest2, err := test.impl.Create("test-pod", "test-owner", nodeID2, validPodSpec)
	test.NoError(err)
	test.Equal("test-pod", digest2.Name)
	test.NotEmpty(digest2.Uuid)
	test.Equal("test-owner", digest2.Owner)
	test.NotEmpty(digest2.State)

	test.NotEqual(digest1.Uuid, digest2.Uuid)

	pod1, err := test.podKvs.Get(digest1.Uuid)
	test.NoError(err)
	test.Equal("test-pod", pod1.Meta.Name)
	test.Equal("test-owner", pod1.Meta.Owner)
	test.Equal(nodeID1, pod1.Meta.CreatorNode)
	test.Equal(digest1.Uuid, pod1.Meta.Uuid)
	test.NotNil(pod1.Spec)

	// get pod data
	pod1, err = test.impl.GetPodData(digest1.Uuid)
	test.NoError(err)
	test.Equal("test-pod", pod1.Meta.Name)
	test.Equal("test-owner", pod1.Meta.Owner)
	test.Equal(nodeID1, pod1.Meta.CreatorNode)
	test.Equal(digest1.Uuid, pod1.Meta.Uuid)
	test.NotNil(pod1.Spec)

	// get pod data with uuid does not exist
	pod0, err := test.impl.GetPodData("not exist")
	test.Error(err)
	test.Nil(pod0)

	// migrate a pod
	test.Empty(pod1.Spec.TargetNode)
	test.Empty(pod1.Status.RunningNode)

	err = test.impl.Migrate(digest1.Uuid, nodeID1)
	test.NoError(err)
	pod1, err = test.podKvs.Get(digest1.Uuid)
	test.NoError(err)
	test.Equal(nodeID1, pod1.Spec.TargetNode)
	test.Equal(nodeID1, pod1.Status.RunningNode)

	err = test.impl.Migrate(digest1.Uuid, nodeID2)
	test.NoError(err)
	pod1, err = test.podKvs.Get(digest1.Uuid)
	test.NoError(err)
	test.Equal(nodeID2, pod1.Spec.TargetNode)
	test.Equal(nodeID1, pod1.Status.RunningNode)

	// delete a pod
	state := "before delete"
	go func() {
		err := test.impl.Delete(digest1.Uuid)
		test.NoError(err)
		state = "deleted"
	}()

	// wait for being added to the deletion timestamp
	test.Eventually(func() bool {
		pod1, err = test.podKvs.Get(digest1.Uuid)
		test.NoError(err)
		return pod1.Meta.DeletionTimestamp != ""
	}, 1*time.Minute, 1*time.Second)

	// DealLocalResource will be delete the record
	test.podKvs.Delete(digest1.Uuid)

	test.Eventually(func() bool {
		return state == "deleted"
	}, 1*time.Minute, 1*time.Second)

	// delete a pod with uuid it delete yet
	err = test.impl.Delete(digest1.Uuid)
	test.NoError(err)

	// cleanup a pod that is deleted yet
	err = test.impl.Cleanup(digest1.Uuid)
	test.Error(err)

	// cleanup a pod with healthy pod
	err = test.impl.Cleanup(digest2.Uuid)
	test.Error(err)
	// cleanup a pod with unhealthy pod
	pod2, err := test.impl.GetPodData(digest2.Uuid)
	test.NoError(err)
	pod2.Status.ContainerStatuses[0].State.Unknown = &api.ContainerStateUnknown{
		Reason:    "test",
		Timestamp: misc.GetTimestamp(),
	}
	err = test.podKvs.Update(pod2)
	test.NoError(err)

	err = test.impl.Cleanup(digest2.Uuid)
	test.NoError(err)

	// pod1 & pod2 get deleted
	_, err = test.podKvs.Get(digest1.Uuid)
	test.Error(err)
	_, err = test.podKvs.Get(digest2.Uuid)
	test.Error(err)
}
