package local

import (
	"github.com/llamerada-jp/oinari/agent/core"
	"github.com/llamerada-jp/oinari/agent/crosslink"
)

type connectParam struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type applyPodParam struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type applyPodResult struct {
	Uuid string `json:"uuid"`
}

type setPositionParam struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type terminateParam struct {
	Uuid string `json:"uuid"`
}

func InitHandler(system *core.System, rootHandler crosslink.MultiPlexer) error {
	mpx := crosslink.NewMultiPlexer()
	rootHandler.SetHandler("system", mpx)

	mpx.SetHandler("connect", crosslink.NewFuncObjHandler(func(param *connectParam, tags map[string]string, writer crosslink.ResponseObjWriter) {
		err := system.Connect(param.Url, param.Token)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("applyPod", crosslink.NewFuncObjHandler(func(param *applyPodParam, tags map[string]string, writer crosslink.ResponseObjWriter) {
		uuid, err := system.ApplyPod(param.Name, param.Image)
		if err != nil {
			writer.ReplyError((err.Error()))
			return
		}
		writer.ReplySuccess(applyPodResult{
			Uuid: uuid,
		})
	}))

	mpx.SetHandler("setPosition", crosslink.NewFuncObjHandler(func(param *setPositionParam, tags map[string]string, writer crosslink.ResponseObjWriter) {
		err := system.SetPosition(param.Latitude, param.Longitude)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("terminate", crosslink.NewFuncObjHandler(func(param *terminateParam, tags map[string]string, writer crosslink.ResponseObjWriter) {
		err := system.Terminate(param.Uuid)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	return nil
}
