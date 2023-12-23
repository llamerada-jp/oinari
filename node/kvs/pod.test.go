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
package kvs

import (
	"encoding/json"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/node/misc"
	"github.com/llamerada-jp/oinari/node/mock"
	"github.com/stretchr/testify/suite"
)

type podKvsTest struct {
	suite.Suite
	col  *mock.Colonio
	impl PodKvs
}

var (
	validSpec = &core.PodSpec{
		Containers: []core.ContainerSpec{
			{
				Name:          "test",
				Image:         "http://localhost/dummy.wasm",
				Runtime:       []string{"go:1.20"},
				RestartPolicy: core.RestartPolicyAlways,
			},
		},
	}

	validStatus = &core.PodStatus{
		RunningNode: "01234567890123456789012345abcdef",
		ContainerStatuses: []core.ContainerStatus{
			{
				ContainerID: "dummy",
				Image:       "http://localhost/dummy.wasm",
				State: core.ContainerState{
					Running: &core.ContainerStateRunning{
						StartedAt: misc.GetTimestamp(),
					},
				},
			},
		},
	}
)

func NewPodKvsTest() suite.TestingSuite {
	colonioMock := mock.NewColonioMock()
	return &podKvsTest{
		col:  colonioMock,
		impl: NewPodKvs(colonioMock),
	}
}

func (test *podKvsTest) TestCreate() {
	uuid := core.GeneratePodUuid()
	name1 := "cat"
	name2 := "dog"
	name3 := "naked mole rat"

	/// normal pattern
	err := test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name1,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)
	test.getByUUID(uuid)

	/// abnormal: duplicate uuid
	err = test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name2,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.Error(err)
	pod := test.getByUUID(uuid)
	test.Equal(name1, pod.Meta.Name)

	/// normal pattern: the entry deleted
	err = test.impl.Delete(uuid)
	test.NoError(err)
	err = test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name2,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)
	pod = test.getByUUID(uuid)
	test.Equal(name2, pod.Meta.Name)

	/// abnormal: failed to validate
	err = test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:  core.ResourceTypePod,
			Name:  name3,
			Owner: "owner",
			// CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.Error(err)
	pod = test.getByUUID(uuid)
	test.Equal(name2, pod.Meta.Name)
}

func (test *podKvsTest) TestUpdate() {
	uuid := core.GeneratePodUuid()
	name1 := "cat"
	name2 := "dog"
	name3 := "naked mole rat"

	/// normal pattern
	err := test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name1,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)

	err = test.impl.Update(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name2,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)
	pod := test.getByUUID(uuid)
	test.Equal(name2, pod.Meta.Name)

	/// abnormal: the entry is not exist
	newUuid := core.GeneratePodUuid()
	err = test.impl.Update(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name2,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              newUuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.Error(err)
	test.False(test.existsByUUID(newUuid))

	/// abnormal: failed to validate
	err = test.impl.Update(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:  core.ResourceTypePod,
			Name:  name3,
			Owner: "owner",
			// CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.Error(err)
	pod = test.getByUUID(uuid)
	test.Equal(name2, pod.Meta.Name)
}

func (test *podKvsTest) TestGet() {
	uuid := core.GeneratePodUuid()
	name := "naked mole rat"
	err := test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              name,
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)

	/// normal pattern
	pod, err := test.impl.Get(uuid)
	test.NoError(err)
	test.Equal(name, pod.Meta.Name)

	/// abnormal: entry not exist
	_, err = test.impl.Get(core.GeneratePodUuid())
	test.Error(err)

	/// abnormal: the entry deleted
	err = test.impl.Delete(uuid)
	test.NoError(err)
	_, err = test.impl.Get(uuid)
	test.Error(err)
}

func (test *podKvsTest) TestDelete() {
	uuid := core.GeneratePodUuid()
	err := test.impl.Create(&core.Pod{
		Meta: &core.ObjectMeta{
			Type:              core.ResourceTypePod,
			Name:              "naked mole rat",
			Owner:             "owner",
			CreatorNode:       "01234567890123456789012345678901",
			Uuid:              uuid,
			DeletionTimestamp: "",
		},
		Spec:   validSpec,
		Status: validStatus,
	})
	test.NoError(err)

	/// normal pattern
	err = test.impl.Delete(uuid)
	test.NoError(err)
	val, err := test.col.KvsGet(string(core.ResourceTypePod) + "/" + uuid)
	test.NoError(err)
	test.True(val.IsNil())

	/// normal pattern: double delete
	err = test.impl.Delete(uuid)
	test.NoError(err)

	/// normal pattern: entry not exit
	err = test.impl.Delete(core.GeneratePodUuid())
	test.NoError(err)
}

func (test *podKvsTest) getByUUID(uuid string) *core.Pod {
	key := string(core.ResourceTypePod) + "/" + uuid
	val, err := test.col.KvsGet(key)
	test.NoError(err)
	raw, err := val.GetBinary()
	test.NoError(err)
	pod := &core.Pod{}
	err = json.Unmarshal(raw, pod)
	test.NoError(err)
	return pod
}

func (test *podKvsTest) existsByUUID(uuid string) bool {
	key := string(core.ResourceTypePod) + "/" + uuid
	_, err := test.col.KvsGet(key)

	if err != nil {
		test.ErrorIs(colonio.ErrKvsNotFound, err)
		return false
	}
	return true
}
