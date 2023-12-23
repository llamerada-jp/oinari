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
package oinari

import (
	"fmt"
	"log"

	"github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type Manager struct {
	// crosslink
	cl      crosslink.Crosslink
	rootMpx crosslink.MultiPlexer
	// apis
	apis []API
	// application
	app Application
}

func NewManager() *Manager {
	return &Manager{
		apis: make([]API, 0),
	}
}

func (m *Manager) Use(api API) error {
	m.apis = append(m.apis, api)
	return nil
}

func (m *Manager) Run(app Application) error {
	m.app = app
	errCh := make(chan error)

	if err := m.init(errCh); err != nil {
		return fmt.Errorf("failed to init the manager: %w", err)
	}

	if err := m.ready(); err != nil {
		return fmt.Errorf("failed to ready the manager: %w", err)
	}

	err, ok := <-errCh
	if ok {
		close(errCh)
		return err
	}
	return nil
}

func (m *Manager) init(errCh chan error) error {
	// setup crosslink
	m.rootMpx = crosslink.NewMultiPlexer()
	m.cl = crosslink.NewCrosslink("crosslink", m.rootMpx)
	appMpx := crosslink.NewMultiPlexer()
	m.rootMpx.SetHandler("application", appMpx)
	apiMpx := crosslink.NewMultiPlexer()
	appMpx.SetHandler("api", apiMpx)

	// setup std writer
	if err := initWriter(m.cl, NodeCrosslinkPath); err != nil {
		return fmt.Errorf("failed to setup std writer module: %w", err)
	}

	// setup APIs
	if err := m.setupCoreHandler(apiMpx, errCh); err != nil {
		return fmt.Errorf("failed to setup core api: %w", err)
	}
	for _, api := range m.apis {
		if err := api.Setup(m.cl, apiMpx, errCh); err != nil {
			return fmt.Errorf("failed to setup an api(%s): %w", api.Name(), err)
		}
	}

	return nil
}

func (m *Manager) setupCoreHandler(apiMpx crosslink.MultiPlexer, errCh chan error) error {
	coreAPIMpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("core", coreAPIMpx)

	coreAPIMpx.SetHandler("setup", crosslink.NewFuncHandler(func(req *core.SetupRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := m.app.Setup(req.IsInitialize, req.Record)
		if err != nil {
			writer.ReplyError("setup had an error")
			errCh <- fmt.Errorf("catch an error on `Setup` method: %s", err)
		} else {
			writer.ReplySuccess(core.SetupResponse{})
		}
	}))

	coreAPIMpx.SetHandler("marshal", crosslink.NewFuncHandler(func(req *core.MarshalRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		record, err := m.app.Marshal()
		if err != nil {
			writer.ReplyError("marshal had an error")
			errCh <- fmt.Errorf("catch an error on `Marshal` method: %s", err)
		} else {
			writer.ReplySuccess(core.MarshalResponse{
				Record: record,
			})
		}
	}))

	coreAPIMpx.SetHandler("teardown", crosslink.NewFuncHandler(func(req *core.TeardownRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		record, err := m.app.Teardown(req.IsFinalize)
		// ignore record if finalize
		if req.IsFinalize {
			record = nil
		}
		if err != nil {
			writer.ReplyError("teardown had an error")
			errCh <- fmt.Errorf("catch an error on `Teardown` method: %s", err)
		} else {
			writer.ReplySuccess(core.TeardownResponse{
				Record: record,
			})
			close(errCh)
		}
	}))

	return nil
}

func (m *Manager) ready() error {
	m.cl.Call(NodeCrosslinkPath+"/ready", core.ReadyRequest{}, nil, func(b []byte, err error) {
		if err != nil {
			log.Fatalf("failed to ready core api of oinari, do you have a runtime set?: %s", err.Error())
		}
	})
	return nil
}
