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
	"fmt"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node"
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

// implement system events
func (na *nodeAgent) OnConnect() error {
	ctx := context.Background()

	// node controller
	localNid := na.col.GetLocalNid()
	nodeCtrl := controller.NewNodeController(localNid)

	// account controller
	accountKvs := kvs.NewAccountKvs(na.col)
	accountCtrl := controller.NewAccountController(ctx, na.sysCtrl.GetAccount(), nodeCtrl.GetNid(), accountKvs)

	// pod controller
	podKvs := kvs.NewPodKvs(na.col)
	podMsg := md.NewMessagingDriver(na.col)
	podCtrl := controller.NewPodController(podKvs, podMsg, localNid)

	// container controller
	cri := cri.NewCRI(na.cl)
	containerCtrl := controller.NewContainerController(localNid, cri, podKvs)

	// manager
	ld := node.NewLocalDatastore(na.col)
	nodeMgr := node.NewManager(ld, podCtrl)
	go func() {
		err := nodeMgr.Start(na.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// handlers
	mh.InitMessagingHandler(containerCtrl, na.col)
	fh.InitResourceHandler(na.rootMpx, accountCtrl, containerCtrl, nodeCtrl, podCtrl)

	return nil
}

func main() {
	na := &nodeAgent{}
	err := na.execute()
	if err != nil {
		fmt.Println(err)
	}
}
