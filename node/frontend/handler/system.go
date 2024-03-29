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
	"log"

	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/controller"
)

type connectRequest struct {
	Url      string `json:"url"`
	Account  string `json:"account"`
	Token    string `json:"token"`
	NodeName string `json:"nodeName"`
	NodeType string `json:"nodeType"`
}

type connectResponse struct {
	Account string `json:"account"`
	NodeID  string `json:"nodeID"`
}

type closeRequest struct {
}

type Info struct {
	CommitHash string `json:"commitHash"`
}

func InitSystemHandler(nodeMpx crosslink.MultiPlexer, appFilter controller.ApplicationFilter, sysCtrl controller.SystemController, info *Info) error {
	mpx := crosslink.NewMultiPlexer()
	nodeMpx.SetHandler("system", mpx)

	mpx.SetHandler("info", crosslink.NewFuncHandler(func(_ *interface{}, tags map[string]string, writer crosslink.ResponseWriter) {
		writer.ReplySuccess(info)
	}))

	mpx.SetHandler("connect", crosslink.NewFuncHandler(func(request *connectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sysCtrl.Connect(request.Url, request.Account, request.Token, request.NodeName, request.NodeType)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(connectResponse{
			Account: sysCtrl.GetAccount(),
			NodeID:  sysCtrl.GetNode(),
		})
	}))

	mpx.SetHandler("disconnect", crosslink.NewFuncHandler(func(param *closeRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sysCtrl.Disconnect()
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("config", crosslink.NewFuncHandler(
		func(param *configRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			switch param.Key {
			case "allowApplications":
				appFilter.SetFilter(param.Value)
			case "samplePrefix":
				appFilter.SetSamplePrefix(param.Value)
			default:
				log.Fatalln("unknown config key: ", param.Key)
			}
			writer.ReplySuccess(nil)
		}))

	return nil
}
