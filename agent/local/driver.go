package local

import (
	"github.com/llamerada-jp/oinari/agent/core"
	"github.com/llamerada-jp/oinari/agent/crosslink"
)

type driverImpl struct {
	cl crosslink.Crosslink
}

func NewCommandDriver(cl crosslink.Crosslink) core.LocalCommandDriver {
	return &driverImpl{
		cl: cl,
	}
}
