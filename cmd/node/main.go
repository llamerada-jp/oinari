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
	"github.com/llamerada-jp/oinari/node/cri"
	"github.com/llamerada-jp/oinari/node/resource"
	"github.com/llamerada-jp/oinari/node/resource/account"
	"github.com/llamerada-jp/oinari/node/resource/node"
	"github.com/llamerada-jp/oinari/node/resource/pod"
	"github.com/llamerada-jp/oinari/node/system"
)

type nodeAgent struct {
	ctx context.Context
	// crosslink
	cl      crosslink.Crosslink
	rootMpx crosslink.MultiPlexer
	// colonio
	col colonio.Colonio
	// system
	sys system.System
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
	cd := system.NewCommandDriver(na.cl)
	na.sys = system.NewSystem(na.col, na, cd)
	go func() {
		err := na.sys.Start(na.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	system.InitCommandHandler(na.sys, na.rootMpx)
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

	err = na.sys.TellInitComplete()
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// implement system events
func (na *nodeAgent) OnConnect() error {
	// node manager
	nodeMgr := node.NewManager(na.col.GetLocalNid())

	// account manager
	accountKvs := account.NewKvsDriver(na.col)
	accountMgr := account.NewManager(na.sys.GetAccount(), nodeMgr.GetNid(), accountKvs)

	// pod manager
	cri := cri.NewCRI(na.cl)
	podKvs := pod.NewKvsDriver(na.col)
	podMsg := pod.NewMessagingDriver(na.col)
	podMgr := pod.NewManager(cri, podKvs, podMsg, accountMgr, nodeMgr)
	pod.InitCommandHandler(podMgr, na.rootMpx)
	pod.InitMessagingHandler(podMgr, na.col)

	// resource manager
	ld := resource.NewLocalDatastore(na.col)
	resourceManager := resource.NewManager(ld, podMgr)
	go func() {
		err := resourceManager.Start(na.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	return nil
}

func main() {
	na := &nodeAgent{}
	err := na.execute()
	if err != nil {
		fmt.Println(err)
	}
}
