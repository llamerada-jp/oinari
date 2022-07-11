package global

import (
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/agent/core"
)

func InitHandler(system *core.System, col colonio.Colonio) error {
	col.OnCall("encouragePod", func(cp *colonio.CallParameter) interface{} {
		js, err := cp.Value.GetString()
		if err != nil {
			log.Println(err)
			return nil
		}
		var param encouragePod
		err = json.Unmarshal([]byte(js), &param)
		if err != nil {
			log.Println(err)
			return nil
		}

		err = system.EncouragePod(context.Background(), param.Uuid)
		if err != nil {
			log.Println(err)
			return nil
		}

		return nil
	})
	return nil
}
