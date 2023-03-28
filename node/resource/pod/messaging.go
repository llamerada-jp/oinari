package pod

import (
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type MessagingDriver interface {
	encouragePod(nid string, pod *api.Pod) error
}

type messagingDriverImpl struct {
	colonio colonio.Colonio
}

type encouragePodMessage struct {
	Pod *api.Pod `json:"pod"`
}

func InitMessagingHandler(podMgr Manager, col colonio.Colonio) error {
	col.MessagingSetHandler("encouragePod", func(mr *colonio.MessagingRequest, mrw colonio.MessagingResponseWriter) {
		raw, err := mr.Message.GetBinary()
		defer mrw.Write(nil)
		if err != nil {
			log.Println(err)
			return
		}
		var msg encouragePodMessage
		err = json.Unmarshal(raw, &msg)
		if err != nil {
			log.Println(err)
			return
		}

		err = podMgr.encouragePod(context.Background(), msg.Pod)
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

func (d *messagingDriverImpl) encouragePod(nid string, pod *api.Pod) error {
	raw, err := json.Marshal(encouragePodMessage{
		Pod: pod,
	})
	if err != nil {
		return err
	}
	_, err = d.colonio.MessagingPost(nid, "encouragePod", raw, 0)
	if err != nil {
		return err
	}
	return nil
}
