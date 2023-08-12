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
	"context"
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/node/frontend/driver"
	"golang.org/x/exp/slices"
)

type EventHandler interface {
	OnConnect(nodeName string, nodeType api.NodeType) error
}

type SystemController interface {
	Start(ctx context.Context) error
	Connect(url, account, token, nodeName, nodeType string) error
	Disconnect() error

	GetAccount() string
	GetNode() string
}

type systemControllerImpl struct {
	colonio        colonio.Colonio
	evh            EventHandler
	frontendDriver driver.FrontendDriver
	account        string
}

func NewSystemController(col colonio.Colonio, evh EventHandler, frontendDriver driver.FrontendDriver) SystemController {
	return &systemControllerImpl{
		colonio:        col,
		evh:            evh,
		frontendDriver: frontendDriver,
	}
}

func (impl *systemControllerImpl) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (impl *systemControllerImpl) GetAccount() string {
	return impl.account
}

func (impl *systemControllerImpl) GetNode() string {
	return impl.colonio.GetLocalNid()
}

func (impl *systemControllerImpl) Connect(url, account, token, nodeName, nodeType string) error {
	if !slices.Contains(api.NodeTypeAccepted, api.NodeType(nodeType)) {
		return fmt.Errorf("unsupported node type specified")
	}

	err := impl.colonio.Connect(url, token)
	if err != nil {
		return fmt.Errorf("failed to colonio.Connect: %w", err)
	}

	impl.account = account

	err = impl.evh.OnConnect(nodeName, api.NodeType(nodeType))
	if err != nil {
		return err
	}

	return nil
}

func (impl *systemControllerImpl) Disconnect() error {
	err := impl.colonio.Disconnect()
	if err != nil {
		return err
	}
	impl.account = ""
	return nil
}
