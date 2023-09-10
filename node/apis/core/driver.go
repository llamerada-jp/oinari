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
package core

import (
	"encoding/json"
	"strings"

	app "github.com/llamerada-jp/oinari/api/app/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/lib/oinari"
)

type coreAPIDriverImpl struct {
	cl          crosslink.Crosslink
	containerID string
	tags        map[string]string
}

func NewCoreAPIDriver(cl crosslink.Crosslink, containerID string) CoreDriver {
	return &coreAPIDriverImpl{
		cl:          cl,
		containerID: containerID,
		tags: map[string]string{
			"containerID": containerID,
		},
	}
}

func (driver *coreAPIDriverImpl) DriverName() string {
	return "core:dev1"
}

func callHelper[REQ any, RES any](driver *coreAPIDriverImpl, path string, request *REQ) (*RES, error) {
	ch := make(chan *RES)
	var funcError error

	driver.cl.Call(strings.Join([]string{oinari.ApplicationCrosslinkPath, path}, "/"), request, nil,
		func(response []byte, err error) {
			defer close(ch)

			if err != nil {
				funcError = err
				return
			}

			var res RES
			err = json.Unmarshal(response, &res)
			if err != nil {
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

func (driver *coreAPIDriverImpl) Setup(firstInPod bool) error {
	_, err := callHelper[app.SetupRequest, app.SetupResponse](driver, "setup", &app.SetupRequest{
		FirstInPod: firstInPod,
	})
	return err
}

func (driver *coreAPIDriverImpl) Dump() ([]byte, error) {
	res, err := callHelper[app.DumpRequest, app.DumpResponse](driver, "dump", &app.DumpRequest{})
	if err != nil {
		return nil, err
	}
	return res.DumpData, nil
}

func (driver *coreAPIDriverImpl) Restore(dumpData []byte) error {
	_, err := callHelper[app.RestoreRequest, app.RestoreResponse](driver, "restore", &app.RestoreRequest{
		DumpData: dumpData,
	})
	return err
}

func (driver *coreAPIDriverImpl) Teardown(lastInPod bool) error {
	_, err := callHelper[app.TeardownRequest, app.TeardownResponse](driver, "teardown", &app.TeardownRequest{
		LastInPod: lastInPod,
	})
	return err
}
