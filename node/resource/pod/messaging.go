package pod

import (
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type MessagingDriver interface {
	vitalizePod(nid, uuid string) error
}

type messagingDriverImpl struct {
	colonio colonio.Colonio
}

type vitalizePodMessage struct {
	Uuid string `json:"uuid"`
}

func InitMessagingHandler(podMgr Manager, col colonio.Colonio) error {
	col.MessagingSetHandler("vitalizePod", func(mr *colonio.MessagingRequest, mrw colonio.MessagingResponseWriter) {
		raw, err := mr.Message.GetBinary()
		defer mrw.Write(nil)
		if err != nil {
			log.Println(err)
			return
		}
		var msg vitalizePodMessage
		err = json.Unmarshal(raw, &msg)
		if err != nil {
			log.Println(err)
			return
		}

		err = podMgr.updateContainer(context.Background(), msg.Uuid)
		if err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}

func NewMessagingDriver(col colonio.Colonio) MessagingDriver {
	return &messagingDriverImpl{
		colonio: col,
	}
}

func (d *messagingDriverImpl) vitalizePod(nid, uuid string) error {
	raw, err := json.Marshal(vitalizePodMessage{
		Uuid: uuid,
	})
	if err != nil {
		return err
	}
	_, err = d.colonio.MessagingPost(nid, "vitalizePod", raw, 0)
	if err != nil {
		return err
	}
	return nil
}
