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
package api

import (
	"testing"

	"github.com/llamerada-jp/oinari/node/misc"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePodUuid(t *testing.T) {
	assert := assert.New(t)

	id1 := GeneratePodUuid()
	assert.NotEmpty(id1)
	id2 := GeneratePodUuid()
	assert.NotEmpty(id2)
	assert.NotEqual(id1, id2)
}

func TestValidatePodUuid(t *testing.T) {
	assert := assert.New(t)

	validUuid := GeneratePodUuid()
	assert.NoError(ValidatePodUuid(validUuid))

	for _, uuid := range []string{
		"",
		validUuid + "z",
	} {
		assert.Error(ValidatePodUuid(uuid))
	}
}

func TestValidatePod(t *testing.T) {
	assert := assert.New(t)

	validMeta := &ObjectMeta{
		Type:              ResourceTypePod,
		Name:              "name",
		Owner:             "owner",
		CreatorNode:       "01234567890123456789012345678901",
		Uuid:              GeneratePodUuid(),
		DeletionTimestamp: "",
	}
	assert.NoError(validMeta.Validate(ResourceTypePod))

	validSpec := &PodSpec{
		Containers: []ContainerSpec{
			{
				Name:    "test",
				Image:   "http://localhost/dummy.wasm",
				Runtime: []string{"go:1.20"},
			},
		},
	}
	assert.NoError(validSpec.validate())

	validStatus := &PodStatus{
		RunningNode: "01234567890123456789012345abcdef",
		ContainerStatuses: []ContainerStatus{
			{
				ContainerID: "dummy",
				Image:       "http://localhost/dummy.wasm",
				State:       ContainerState{},
			},
		},
	}
	assert.NoError(validStatus.validate(1))

	// valid
	for _, tc := range []struct {
		mustStatus bool
		pod        *Pod
	}{
		{
			mustStatus: false,
			pod: &Pod{
				Meta:   validMeta,
				Spec:   validSpec,
				Status: validStatus,
			},
		},
		{
			mustStatus: true,
			pod: &Pod{
				Meta:   validMeta,
				Spec:   validSpec,
				Status: validStatus,
			},
		},
		// no status without status check
		{
			mustStatus: false,
			pod: &Pod{
				Meta: validMeta,
				Spec: validSpec,
			},
		},
	} {
		assert.NoError(tc.pod.Validate(tc.mustStatus))
	}

	// invalid
	for title, tc := range map[string]struct {
		mustStatus bool
		pod        *Pod
	}{
		"no status with mustStatus": {
			mustStatus: true,
			pod: &Pod{
				Meta: validMeta,
				Spec: validSpec,
			},
		},
		"invalid uuid": {
			mustStatus: false,
			pod: &Pod{
				Meta: &ObjectMeta{
					Type:              ResourceTypePod,
					Name:              "name",
					Owner:             "owner",
					CreatorNode:       "01234567890123456789012345678901",
					Uuid:              GeneratePodUuid() + " invalid",
					DeletionTimestamp: "",
				},
				Spec: validSpec,
			},
		},
		"invalid meta": {
			mustStatus: false,
			pod: &Pod{
				Meta:   &ObjectMeta{},
				Spec:   validSpec,
				Status: validStatus,
			},
		},
		"invalid spec": {

			mustStatus: false,
			pod: &Pod{
				Meta:   validMeta,
				Spec:   &PodSpec{},
				Status: validStatus,
			},
		},
		"invalid status": {
			mustStatus: false,
			pod: &Pod{
				Meta:   validMeta,
				Spec:   validSpec,
				Status: &PodStatus{},
			},
		},
	} {
		assert.Error(tc.pod.Validate(tc.mustStatus), title)
	}
}

func TestValidatePodSpec(t *testing.T) {
	assert := assert.New(t)

	// valid
	for title, spec := range map[string]*PodSpec{
		"single container all filled": {
			Containers: []ContainerSpec{
				{
					Name:    "test",
					Image:   "http://localhost/test.wasm",
					Runtime: []string{"go:1.20"},
					Args:    []string{"test arg"},
					Env: []EnvVar{
						{
							Name:  "key",
							Value: "val",
						},
					},
				},
			},
		},
		"multi container": {
			Containers: []ContainerSpec{
				{
					Name:    "c1",
					Image:   "http://localhost/test1.wasm",
					Runtime: []string{"go:1.20"},
				},
				{
					Name:    "c2",
					Image:   "http://localhost/test2.wasm",
					Runtime: []string{"go:1.19"},
				},
			},
		},
	} {
		assert.NoError(spec.validate(), title)
	}

	// invalid
	for title, spec := range map[string]*PodSpec{
		"nil": nil,
		"no container": {
			Containers: []ContainerSpec{},
		},
		"no container name": {
			Containers: []ContainerSpec{
				{
					// Name:    "test",
					Image:   "http://localhost/test.wasm",
					Runtime: []string{"go:1.20"},
				},
			},
		},
		"duplicate container name": {
			Containers: []ContainerSpec{
				{
					Name:    "test",
					Image:   "http://localhost/test1.wasm",
					Runtime: []string{"go:1.20"},
				},
				{
					Name:    "test",
					Image:   "http://localhost/test2.wasm",
					Runtime: []string{"go:1.19"},
				},
			},
		},
		"no image": {
			Containers: []ContainerSpec{
				{
					Name: "test",
					// Image:   "http://localhost/test.wasm",
					Runtime: []string{"go:1.20"},
				},
			},
		},
		"invalid runtime": {
			Containers: []ContainerSpec{
				{
					Name:    "test",
					Image:   "http://localhost/test.wasm",
					Runtime: []string{"??"},
				},
			},
		},
		"duplicate env key": {
			Containers: []ContainerSpec{
				{
					Name:    "test",
					Image:   "http://localhost/test.wasm",
					Runtime: []string{"go:1.20"},
					Args:    []string{"test arg"},
					Env: []EnvVar{
						{
							Name:  "key",
							Value: "val1",
						},
						{
							Name:  "key",
							Value: "val2",
						},
					},
				},
			},
		},
	} {
		assert.Error(spec.validate(), title)
	}
}

func TestValidatePodStatus(t *testing.T) {
	assert := assert.New(t)

	for _, status := range []*PodStatus{
		{
			ContainerStatuses: []ContainerStatus{
				{},
			},
		},
		{
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{},
			},
		},
		{
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		{
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
						Terminated: &ContainerStateTerminated{
							FinishedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
	} {
		assert.NoError(status.validate(1))
	}

	for title, status := range map[string]*PodStatus{
		"container state is not specified": {
			RunningNode: "01234567890123456789012345abcdef",
		},
		"invalid node id": {
			RunningNode: "no no no",
			ContainerStatuses: []ContainerStatus{
				{},
			},
		},
		"only running node specified": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{},
			},
		},
		"invalid container id field": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					// ContainerID: "container1",
					Image: "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		"invalid image field": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					// Image: "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		"invalid running field": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State:       ContainerState{},
				},
			},
		},
		"invalid startedAt field": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							// StartedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		"running should be set if terminated is set": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: nil,
						Terminated: &ContainerStateTerminated{
							FinishedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		"finishedAt required when terminated is set": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
						Terminated: &ContainerStateTerminated{
							// FinishedAt: misc.GetTimestamp(),
						},
					},
				},
			},
		},
		"timestamp required when unknown is set": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
						Unknown: &ContainerStateUnknown{
							// Timestamp: misc.GetTimestamp(),
							Reason: "the reason causes unknown status",
						},
					},
				},
			},
		},
		"reason required when unknown is set": {
			RunningNode: "01234567890123456789012345abcdef",
			ContainerStatuses: []ContainerStatus{
				{
					ContainerID: "container1",
					Image:       "https://localhost/dummy.wasm",
					State: ContainerState{
						Running: &ContainerStateRunning{
							StartedAt: misc.GetTimestamp(),
						},
						Unknown: &ContainerStateUnknown{
							Timestamp: misc.GetTimestamp(),
							// Reason: "enjoy!",
						},
					},
				},
			},
		},
	} {
		assert.Error(status.validate(1), title)
	}
}
