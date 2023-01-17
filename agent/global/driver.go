package global

import (
	"encoding/json"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/agent/core"
)

type driverImpl struct {
	colonio colonio.Colonio
}

func NewCommandDriver(col colonio.Colonio) core.GlobalCommandDriver {
	return &driverImpl{
		colonio: col,
	}
}

func (d *driverImpl) EncouragePod(nid, uuid string) error {
	js, err := json.Marshal(encouragePod{
		Uuid: uuid,
	})
	if err != nil {
		return err
	}
	_, err = d.colonio.MessagingPost(nid, "encouragePod", string(js), 0)
	if err != nil {
		return err
	}
	return nil
}
