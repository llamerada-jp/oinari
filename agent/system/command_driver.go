package system

import (
	"log"

	"github.com/llamerada-jp/oinari/agent/crosslink"
)

type CommandDriver interface {
	// send a message that tell initialization complete
	TellInitComplete() error
}

type commandDriver struct {
	cl crosslink.Crosslink
}

func NewCommandDriver(cl crosslink.Crosslink) CommandDriver {
	return &commandDriver{
		cl: cl,
	}
}

func (cd *commandDriver) TellInitComplete() error {
	cd.cl.Call("system/onInitComplete", nil, nil,
		func(result []byte, err error) {
			if err != nil {
				log.Fatalln(err)
			}
		})
	return nil
}
