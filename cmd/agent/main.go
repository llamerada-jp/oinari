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
	"github.com/llamerada-jp/oinari/agent/cri"
	"github.com/llamerada-jp/oinari/agent/resource"
	"github.com/llamerada-jp/oinari/agent/resource/account"
	"github.com/llamerada-jp/oinari/agent/resource/node"
	"github.com/llamerada-jp/oinari/agent/resource/pod"
	"github.com/llamerada-jp/oinari/agent/system"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type agent struct {
	ctx context.Context
	// crosslink
	cl      crosslink.Crosslink
	rootMpx crosslink.MultiPlexer
	// colonio
	col colonio.Colonio
	// system
	sys system.System
}

func (agent *agent) initCrosslink() error {
	agent.rootMpx = crosslink.NewMultiPlexer()
	agent.cl = crosslink.NewCrosslink("crosslink", agent.rootMpx)
	return nil
}

func (agent *agent) initColonio() error {
	config := colonio.NewConfig()
	col, err := colonio.NewColonio(config)
	if err != nil {
		return err
	}
	agent.col = col
	return nil
}

func (agent *agent) initSystem() error {
	cd := system.NewCommandDriver(agent.cl)
	agent.sys = system.NewSystem(agent.col, agent, cd)
	go func() {
		err := agent.sys.Start(agent.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	system.InitCommandHandler(agent.sys, agent.rootMpx)
	return nil
}

func (agent *agent) execute() error {
	ctx, _ := context.WithCancel(context.Background())
	agent.ctx = ctx

	err := agent.initCrosslink()
	if err != nil {
		return err
	}

	err = agent.initColonio()
	if err != nil {
		return err
	}

	err = agent.initSystem()
	if err != nil {
		return err
	}

	err = agent.sys.TellInitComplete()
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// implement system events
func (agent *agent) OnConnect() error {
	// node manager
	nodeMgr := node.NewManager(agent.col.GetLocalNid())

	// account manager
	accountKvs := account.NewKvsDriver(agent.col)
	accountMgr := account.NewManager(agent.sys.GetAccount(), accountKvs)

	// pod manager
	cri := cri.NewCRI(agent.cl)
	podKvs := pod.NewKvsDriver(agent.col)
	podMsg := pod.NewMessagingDriver(agent.col)
	podMgr := pod.NewManager(cri, podKvs, podMsg, accountMgr, nodeMgr)
	pod.InitCommandHandler(podMgr, agent.rootMpx)
	pod.InitMessagingHandler(podMgr, agent.col)

	// resource manager
	ld := resource.NewLocalDatastore(agent.col)
	resourceManager := resource.NewManager(ld, podMgr)
	go func() {
		err := resourceManager.Start(agent.ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	return nil
}

func main() {
	agent := &agent{}
	err := agent.execute()
	if err != nil {
		fmt.Println(err)
	}
}
