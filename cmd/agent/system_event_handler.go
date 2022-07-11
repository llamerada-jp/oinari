package main

import (
	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/agent/core"
	"github.com/llamerada-jp/oinari/agent/global"
)

type systemEventHandlerImpl struct {
	colonio colonio.Colonio
}

func newSystemEventHandler(col colonio.Colonio) core.SystemEventHandler {
	return &systemEventHandlerImpl{
		colonio: col,
	}
}

func (s *systemEventHandlerImpl) OnConnect(sys *core.System) error {
	return global.InitHandler(sys, s.colonio)
}
