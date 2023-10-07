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
package driver

import (
	"encoding/json"
	"fmt"
	"log"

	threeAPI "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type PutObjectRequest struct {
	Objects []threeAPI.Object `json:"objects"`
}

type DeleteObjectRequest struct {
	UUIDs []string `json:"uuids"`
}

type FrontendDriver interface {
	// send a message that tell initialization complete
	TellInitComplete() error
	PutObjects(objects []threeAPI.Object) error
	DeleteObjects(uuids []string) error
}

type frontendDriverImpl struct {
	cl crosslink.Crosslink
}

func NewFrontendDriver(cl crosslink.Crosslink) FrontendDriver {
	return &frontendDriverImpl{
		cl: cl,
	}
}

func (impl *frontendDriverImpl) TellInitComplete() error {
	impl.cl.Call("frontend/nodeReady", nil, nil,
		func(_ []byte, err error) {
			if err != nil {
				log.Fatalf("frontend/nodeReady has an error: %s", err.Error())
			}
		})
	return nil
}

func (impl *frontendDriverImpl) PutObjects(objects []threeAPI.Object) error {
	raw, err := json.Marshal(PutObjectRequest{
		Objects: objects,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	impl.cl.Call("frontend/putObjects", raw, nil, func(b []byte, err error) {
		log.Fatalf("frontend/putObjects has an error: %s", err.Error())
	})
	return nil
}

func (impl *frontendDriverImpl) DeleteObjects(uuids []string) error {
	raw, err := json.Marshal(DeleteObjectRequest{
		UUIDs: uuids,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	impl.cl.Call("frontend/deleteObjects", raw, nil, func(b []byte, err error) {
		log.Fatalf("frontend/deleteObjects has an error: %s", err.Error())
	})
	return nil
}
