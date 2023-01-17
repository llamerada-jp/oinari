package global

import (
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/agent/core"
)

func InitHandler(system *core.System, col colonio.Colonio) error {
	col.MessagingSetHandler("encouragePod", func(mr *colonio.MessagingRequest, mrw colonio.MessagingResponseWriter) {
		js, err := mr.Message.GetString()
		defer mrw.Write(nil)
		if err != nil {
			log.Println(err)
			return
		}
		var param encouragePod
		err = json.Unmarshal([]byte(js), &param)
		if err != nil {
			log.Println(err)
			return
		}

		err = system.EncouragePod(context.Background(), param.Uuid)
		if err != nil {
			log.Println(err)
			return
		}
	})
	return nil
}
