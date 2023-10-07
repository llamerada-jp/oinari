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
package three

import (
	"fmt"
	"log"

	coreAPI "github.com/llamerada-jp/oinari/api/core"
	threeAPI "github.com/llamerada-jp/oinari/api/three"
	coreController "github.com/llamerada-jp/oinari/node/controller"
	kvs "github.com/llamerada-jp/oinari/node/kvs/three"
	messaging "github.com/llamerada-jp/oinari/node/messaging/three/driver"
)

type ObjectController interface {
	Create(name string, podUUID string, spec *threeAPI.ObjectSpec) (string, error)
	Update(uuid string, podUUID string, spec *threeAPI.ObjectSpec) error
	Get(uuid string) (*threeAPI.Object, error)
	Delete(uuid string, podUUID string) error
}

type objectControllerImpl struct {
	// KVSs
	objectKVS kvs.ObjectKVS
	// messaging drivers
	messagingDriver messaging.ThreeMessagingDriver
	// controllers
	nodeCtrl coreController.NodeController
	podCtrl  coreController.PodController
}

func NewObjectController(objectKVS kvs.ObjectKVS, messagingDriver messaging.ThreeMessagingDriver, nodeCtrl coreController.NodeController, podCtrl coreController.PodController) ObjectController {
	return &objectControllerImpl{
		objectKVS:       objectKVS,
		messagingDriver: messagingDriver,
		nodeCtrl:        nodeCtrl,
		podCtrl:         podCtrl,
	}
}

func (impl *objectControllerImpl) Create(name string, podUUID string, spec *threeAPI.ObjectSpec) (string, error) {
	pod, err := impl.podCtrl.GetPodData(podUUID)
	if err != nil {
		return "", fmt.Errorf("failed to get pod data: %w", err)
	}

	obj := &threeAPI.Object{
		Meta: &coreAPI.ObjectMeta{
			Type:        threeAPI.ResourceTypeThreeObject,
			Name:        name,
			Owner:       pod.Meta.Owner,
			CreatorNode: impl.nodeCtrl.GetNid(),
			Uuid:        threeAPI.GenerateObjectUUID(),
		},
		Spec: spec,
	}

	if err := impl.objectKVS.Create(obj); err != nil {
		return "", fmt.Errorf("failed to create object: %w", err)
	}

	if err := impl.messagingDriver.SpreadObject(obj.Meta.Uuid, 10, spec.Position.X, spec.Position.Y); err != nil {
		log.Printf("failed to spread object: name=%s, uuid=%s: %s\n", name, obj.Meta.Uuid, err.Error())
	}

	return obj.Meta.Uuid, nil
}

func (impl *objectControllerImpl) Update(uuid string, podUUID string, spec *threeAPI.ObjectSpec) error {
	pod, err := impl.podCtrl.GetPodData(podUUID)
	if err != nil {
		return fmt.Errorf("failed to get pod data: %w", err)
	}

	obj, err := impl.objectKVS.Get(uuid)
	if err != nil || obj == nil {
		return fmt.Errorf("failed to get object data: %w", err)
	}

	if obj.Meta.Owner != pod.Meta.Owner {
		return fmt.Errorf("object is not owned by pod owner")
	}

	obj.Spec = spec
	if err := impl.objectKVS.Update(obj); err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	if err := impl.messagingDriver.SpreadObject(obj.Meta.Uuid, 10, spec.Position.X, spec.Position.Y); err != nil {
		log.Printf("failed to spread object: name=%s, uuid=%s: %s\n", obj.Meta.Name, obj.Meta.Uuid, err.Error())
	}

	return nil
}

func (impl *objectControllerImpl) Get(uuid string) (*threeAPI.Object, error) {
	obj, err := impl.objectKVS.Get(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return obj, nil
}

func (impl *objectControllerImpl) Delete(uuid string, podUUID string) error {
	pod, err := impl.podCtrl.GetPodData(podUUID)
	if err != nil {
		return fmt.Errorf("failed to get pod data: %w", err)
	}

	obj, err := impl.objectKVS.Get(uuid)
	if err != nil {
		return fmt.Errorf("failed to get object data: %w", err)
	}

	if obj.Meta.Owner != pod.Meta.Owner {
		return fmt.Errorf("object is not owned by pod owner")
	}

	if err := impl.objectKVS.Delete(uuid); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	if err := impl.messagingDriver.SpreadObject(obj.Meta.Uuid, 10*2, obj.Spec.Position.X, obj.Spec.Position.Y); err != nil {
		log.Printf("failed to spread object: name=%s, uuid=%s: %s\n", obj.Meta.Name, obj.Meta.Uuid, err.Error())
	}

	return nil
}
