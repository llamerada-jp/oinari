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
	"math"
	"math/rand"
	"time"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/node/frontend/driver"
)

type EventHandler interface {
	OnConnect() error
}

type SystemController interface {
	Start(ctx context.Context) error
	Connect(url, account, token string) error
	Disconnect() error

	GetAccount() string
	GetNode() string
	SetPosition(latitude, longitude float64) error
}

type systemControllerImpl struct {
	colonio        colonio.Colonio
	evh            EventHandler
	frontendDriver driver.FrontendDriver
	account        string
}

func init() {
	rand.Seed(time.Now().UnixMicro())
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

func (impl *systemControllerImpl) Connect(url, account, token string) error {
	err := impl.colonio.Connect(url, token)
	if err != nil {
		return err
	}

	impl.account = account

	err = impl.evh.OnConnect()
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

func (impl *systemControllerImpl) SetPosition(latitude, longitude float64) error {
	// convert L/L to radian
	_, _, err := impl.colonio.SetPosition(longitude*math.Pi/180.0, latitude*math.Pi/180.0)
	return err
}
