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
package handler

import (
	"github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	kvs "github.com/llamerada-jp/oinari/node/kvs/three"
)

func InitHandler(apiMpx crosslink.MultiPlexer, objectKVS kvs.ObjectKVS) {
	mpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("three", mpx)

	// CreateObject
	mpx.SetHandler("createObject", crosslink.NewFuncHandler(func(request *three.CreateObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		writer.ReplySuccess(&three.CreateObjectResponse{})
	}))

	// UpdateObject
	mpx.SetHandler("updateObject", crosslink.NewFuncHandler(func(request *three.UpdateObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		writer.ReplySuccess(&three.UpdateObjectResponse{})
	}))

	// DeleteObject
	mpx.SetHandler("deleteObject", crosslink.NewFuncHandler(func(request *three.DeleteObjectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		writer.ReplySuccess(&three.DeleteObjectResponse{})
	}))
}
