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
	"encoding/json"
	"fmt"
	"strings"

	api "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type threeAPIImpl struct {
	cl crosslink.Crosslink
}

var _ ThreeAPI = (*threeAPIImpl)(nil)

func NewThreeAPI() ThreeAPI {
	return &threeAPIImpl{}
}

func (impl *threeAPIImpl) Name() string {
	return "three"
}

func (impl *threeAPIImpl) Setup(cl crosslink.Crosslink, apiMpx crosslink.MultiPlexer, errCh chan error) error {
	impl.cl = cl
	return nil
}

func (impl *threeAPIImpl) CreateObject(name string, spec *api.ObjectSpec) (string, error) {
	res, err := callHelper[api.CreateObjectRequest, api.CreateObjectResponse](impl, "createObject", &api.CreateObjectRequest{
		Name: name,
		Spec: spec,
	})
	if err != nil {
		return "", fmt.Errorf("error on CreateObject API: %w", err)
	}
	return res.UUID, nil
}

func (impl *threeAPIImpl) UpdateObject(uuid string, spec *api.ObjectSpec) error {
	_, err := callHelper[api.UpdateObjectRequest, api.UpdateObjectResponse](impl, "updateObject", &api.UpdateObjectRequest{
		UUID: uuid,
		Spec: spec,
	})
	if err != nil {
		return fmt.Errorf("error on UpdateObject API: %w", err)
	}
	return nil
}

func (impl *threeAPIImpl) DeleteObject(uuid string) error {
	_, err := callHelper[api.DeleteObjectRequest, api.DeleteObjectResponse](impl, "deleteObject", &api.DeleteObjectRequest{
		UUID: uuid,
	})
	if err != nil {
		return fmt.Errorf("error on DeleteObject API: %w", err)
	}
	return nil
}

func callHelper[REQ any, RES any](impl *threeAPIImpl, path string, request *REQ) (*RES, error) {
	ch := make(chan *RES)
	var funcError error

	impl.cl.Call(strings.Join([]string{NodeCrosslinkPath, path}, "/"), request, nil,
		func(response []byte, err error) {
			defer close(ch)

			if err != nil {
				funcError = err
				return
			}

			var res RES
			if err := json.Unmarshal(response, &res); err != nil {
				funcError = err
				return
			}

			ch <- &res
		})

	res, ok := <-ch
	if !ok {
		return nil, funcError
	}

	return res, nil
}
