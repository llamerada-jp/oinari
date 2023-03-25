package system

import "github.com/llamerada-jp/oinari/agent/crosslink"

type connectRequest struct {
	Url     string `json:"url"`
	Account string `json:"account"`
	Token   string `json:"token"`
}

type setPositionRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func InitCommandHandler(sys System, rootMpx crosslink.MultiPlexer) error {
	mpx := crosslink.NewMultiPlexer()
	rootMpx.SetHandler("system", mpx)

	mpx.SetHandler("connect", crosslink.NewFuncHandler(func(request *connectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sys.connect(request.Url, request.Account, request.Token)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	mpx.SetHandler("setPosition", crosslink.NewFuncHandler(func(request *setPositionRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sys.setPosition(request.Latitude, request.Longitude)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	return nil
}
