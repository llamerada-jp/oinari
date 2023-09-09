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
package main

import (
	"context"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node"
	"github.com/llamerada-jp/oinari/node/apis"
	"github.com/llamerada-jp/oinari/node/apis/core"
	"github.com/llamerada-jp/oinari/node/apis/null"
	"github.com/llamerada-jp/oinari/node/controller"
	"github.com/llamerada-jp/oinari/node/cri"
	fd "github.com/llamerada-jp/oinari/node/frontend/driver"
	fh "github.com/llamerada-jp/oinari/node/frontend/handler"
	"github.com/llamerada-jp/oinari/node/kvs"
	md "github.com/llamerada-jp/oinari/node/messaging/driver"
	mh "github.com/llamerada-jp/oinari/node/messaging/handler"
)

type nodeAgent struct {
	ctx context.Context
	// crosslink
	cl      crosslink.Crosslink
	rootMpx crosslink.MultiPlexer
	// frontend driver
	frontendDriver fd.FrontendDriver

	// colonio
	col colonio.Colonio
	// system
	sysCtrl controller.SystemController
}

func (na *nodeAgent) initCrosslink() error {
	na.rootMpx = crosslink.NewMultiPlexer()
	na.cl = crosslink.NewCrosslink("crosslink", na.rootMpx)
	return nil
}

func (na *nodeAgent) initColonio() error {
	config := colonio.NewConfig()
	col, err := colonio.NewColonio(config)
	if err != nil {
		return err
	}
	na.col = col
	return nil
}

func (na *nodeAgent) initSystem() error {
	na.frontendDriver = fd.NewFrontendDriver(na.cl)
	na.sysCtrl = controller.NewSystemController(na.col, na, na.frontendDriver)
	go func() {
		err := na.sysCtrl.Start(na.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	fh.InitSystemHandler(na.rootMpx, na.sysCtrl)
	return nil
}

func (na *nodeAgent) execute() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	na.ctx = ctx

	err := na.initCrosslink()
	if err != nil {
		return err
	}

	err = na.initColonio()
	if err != nil {
		return err
	}

	err = na.initSystem()
	if err != nil {
		return err
	}

	err = na.frontendDriver.TellInitComplete()
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func apiDriverFactory(runtime []string) apis.Driver {
	for _, r := range runtime {
		switch r {
		case "core:dev1":
			return core.NewCoreAPIDriver()
		}
	}

	return null.NewNullAPIDriver()
}

// implement system events
func (na *nodeAgent) OnConnect(nodeName string, nodeType api.NodeType) error {
	ctx := context.Background()

	account := na.sysCtrl.GetAccount()
	localNid := na.col.GetLocalNid()

	// CRI
	cri := cri.NewCRI(na.cl)

	// messaging
	messaging := md.NewMessagingDriver(na.col)

	// KVS
	accountKvs := kvs.NewAccountKvs(na.col)
	podKvs := kvs.NewPodKvs(na.col)

	// controllers
	accountCtrl := controller.NewAccountController(account, localNid, accountKvs)
	containerCtrl := controller.NewContainerController(localNid, cri, podKvs, apiDriverFactory)
	nodeCtrl := controller.NewNodeController(ctx, na.col, messaging, account, nodeName, nodeType)
	podCtrl := controller.NewPodController(podKvs, messaging, localNid)

	// manager
	localDs := node.NewLocalDatastore(na.col)
	manager := node.NewManager(localDs, accountCtrl, containerCtrl, nodeCtrl, podCtrl)
	go func() {
		err := manager.Start(na.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// handlers
	mh.InitMessagingHandler(na.col, containerCtrl, nodeCtrl)
	fh.InitResourceHandler(na.rootMpx, accountCtrl, containerCtrl, nodeCtrl, podCtrl)

	return nil
}

func main() {
	na := &nodeAgent{}
	err := na.execute()
	if err != nil {
		log.Println(err)
	}
}
