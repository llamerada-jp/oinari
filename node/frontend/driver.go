package frontend

import (
	"log"

	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type Driver interface {
	// send a message that tell initialization complete
	TellInitComplete() error
}

type driver struct {
	cl crosslink.Crosslink
}

func NewDriver(cl crosslink.Crosslink) Driver {
	return &driver{
		cl: cl,
	}
}

func (cd *driver) TellInitComplete() error {
	cd.cl.Call("frontend/onInitComplete", nil, nil,
		func(_ []byte, err error) {
			if err != nil {
				log.Fatalln(err)
			}
		})
	return nil
}
