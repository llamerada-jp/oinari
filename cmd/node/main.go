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
	api "github.com/llamerada-jp/oinari/api/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node"
	"github.com/llamerada-jp/oinari/node/apis/core"
	ch "github.com/llamerada-jp/oinari/node/apis/core/handler"
	th "github.com/llamerada-jp/oinari/node/apis/three/handler"
	"github.com/llamerada-jp/oinari/node/controller"
	threeController "github.com/llamerada-jp/oinari/node/controller/three"
	"github.com/llamerada-jp/oinari/node/cri"
	fd "github.com/llamerada-jp/oinari/node/frontend/driver"
	fh "github.com/llamerada-jp/oinari/node/frontend/handler"
	coreKVS "github.com/llamerada-jp/oinari/node/kvs"
	threeKVS "github.com/llamerada-jp/oinari/node/kvs/three"
	cmd "github.com/llamerada-jp/oinari/node/messaging/driver"
	cmh "github.com/llamerada-jp/oinari/node/messaging/handler"
	tmd "github.com/llamerada-jp/oinari/node/messaging/three/driver"
	tmh "github.com/llamerada-jp/oinari/node/messaging/three/handler"
)

type nodeAgent struct {
	ctx context.Context
	// crosslink
	cl      crosslink.Crosslink
	nodeMpx crosslink.MultiPlexer
	apiMpx  crosslink.MultiPlexer

	// frontend driver
	frontendDriver fd.FrontendDriver

	// colonio
	col colonio.Colonio
	// system
	sysCtrl   controller.SystemController
	appFilter controller.ApplicationFilter
}

func (na *nodeAgent) initCrosslink() error {
	rootMpx := crosslink.NewMultiPlexer()
	na.nodeMpx = crosslink.NewMultiPlexer()
	rootMpx.SetHandler("node", na.nodeMpx)
	na.apiMpx = crosslink.NewMultiPlexer()
	na.nodeMpx.SetHandler("api", na.apiMpx)
	na.cl = crosslink.NewCrosslink("crosslink", rootMpx)
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
	na.appFilter = controller.NewApplicationFilter()

	go func() {
		err := na.sysCtrl.Start(na.ctx)
		if err != nil {
			log.Fatalf("system controller failed: %s", err.Error())
		}
	}()

	fh.InitSystemHandler(na.nodeMpx, na.appFilter, na.sysCtrl)
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

// implement system events
func (na *nodeAgent) OnConnect(nodeName string, nodeType api.NodeType) error {
	ctx := context.Background()

	account := na.sysCtrl.GetAccount()
	localNid := na.col.GetLocalNid()

	// application filter
	na.appFilter.SetAccount(account)

	// CRI
	cri := cri.NewCRI(na.cl)

	// messaging
	messaging := cmd.NewMessagingDriver(na.col)
	threeMessaging := tmd.NewThreeMessagingDriver(na.col)

	// KVS
	accountKvs := coreKVS.NewAccountKvs(na.col)
	podKvs := coreKVS.NewPodKvs(na.col)
	recordKVS := coreKVS.NewRecordKvs(na.col)
	objectKVS := threeKVS.NewObjectKVS(na.col)

	// api driver manager
	coreDriverManager := core.NewCoreDriverManager(na.cl)

	// controllers
	accountCtrl := controller.NewAccountController(account, localNid, accountKvs)
	containerCtrl := controller.NewContainerController(localNid, cri, na.appFilter, podKvs, recordKVS, coreDriverManager)
	nodeCtrl := controller.NewNodeController(ctx, na.col, messaging, account, nodeName, nodeType)
	podCtrl := controller.NewPodController(podKvs, messaging, localNid)
	objectCtrl := threeController.NewObjectController(objectKVS, na.frontendDriver, threeMessaging, nodeCtrl, podCtrl)

	// manager
	localDs := node.NewLocalDatastore(na.col)
	manager := node.NewManager(localDs, accountCtrl, containerCtrl, nodeCtrl, podCtrl)
	go func() {
		err := manager.Start(na.ctx)
		if err != nil {
			log.Fatalf("node manager failed: %s", err.Error())
		}
	}()

	// handlers
	cmh.InitMessagingHandler(na.col, containerCtrl, nodeCtrl)
	tmh.InitMessagingHandler(na.col, objectCtrl)
	fh.InitResourceHandler(na.nodeMpx, accountCtrl, containerCtrl, nodeCtrl, podCtrl)
	ch.InitHandler(na.apiMpx, coreDriverManager, cri, podKvs, recordKVS)
	th.InitHandler(na.apiMpx, objectCtrl)

	return nil
}

func main() {
	na := &nodeAgent{}
	err := na.execute()
	if err != nil {
		log.Fatalf("error on node: %s", err.Error())
	}
}
