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
	"fmt"
	"log"

	app "github.com/llamerada-jp/oinari/api/app/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/apis/core"
)

func InitHandler(apiMpx crosslink.MultiPlexer, manager *core.Manager) {
	mpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("core", mpx)

	mpx.SetHandler("ready", crosslink.NewFuncHandler(func(request *app.ReadyRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		if !isDriverExists(tags, manager) {
			writer.ReplyError("driver not assigned")
			return
		}
		writer.ReplySuccess(&app.ReadyResponse{})
	}))

	mpx.SetHandler("output", crosslink.NewFuncHandler(func(request *app.OutputRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		if !isDriverExists(tags, manager) {
			writer.ReplyError("driver not assigned")
			return
		}

		// TODO: broadcast message to neighbors
		_, err := fmt.Println(string(request.Payload))
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(&app.OutputResponse{
			Length: len(request.Payload),
		})
	}))
}

func isDriverExists(tags map[string]string, manager *core.Manager) bool {
	containerID, ok := tags["containerID"]
	if !ok {
		log.Fatal("containerID should be set when accessing core handler")
	}

	driver := manager.GetDriver(containerID)
	if driver == nil {
		return false
	}

	if driver.DriverName() == "" {
		return false
	}

	return true
}
