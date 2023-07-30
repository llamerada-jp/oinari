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
package command

import (
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/controller"
)

type connectRequest struct {
	Url     string `json:"url"`
	Account string `json:"account"`
	Token   string `json:"token"`
}

type connectResponse struct {
	Account string `json:"account"`
	Node    string `json:"node"`
}

type setPositionRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func InitSystemHandler(rootMpx crosslink.MultiPlexer, sysCtrl controller.SystemController) error {
	mpx := crosslink.NewMultiPlexer()
	rootMpx.SetHandler("system", mpx)

	mpx.SetHandler("connect", crosslink.NewFuncHandler(func(request *connectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sysCtrl.Connect(request.Url, request.Account, request.Token)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(connectResponse{
			Account: sysCtrl.GetAccount(),
			Node:    sysCtrl.GetNode(),
		})
	}))

	mpx.SetHandler("setPosition", crosslink.NewFuncHandler(func(request *setPositionRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sysCtrl.SetPosition(request.Latitude, request.Longitude)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	return nil
}
