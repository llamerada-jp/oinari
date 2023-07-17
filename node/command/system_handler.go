package command

import (
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/system"
)

type connectRequest struct {
	Url     string `json:"url"`
	Account string `json:"account"`
	Token   string `json:"token"`
}

type connectResponse struct {
	Account string `json:"account"`
	Node    string `json:"node"`
}

type setPositionRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func InitSystemHandler(rootMpx crosslink.MultiPlexer, sys system.System) error {
	mpx := crosslink.NewMultiPlexer()
	rootMpx.SetHandler("system", mpx)

	mpx.SetHandler("connect", crosslink.NewFuncHandler(func(request *connectRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sys.Connect(request.Url, request.Account, request.Token)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(connectResponse{
			Account: sys.GetAccount(),
			Node:    sys.GetNode(),
		})
	}))

	mpx.SetHandler("setPosition", crosslink.NewFuncHandler(func(request *setPositionRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := sys.SetPosition(request.Latitude, request.Longitude)
		if err != nil {
			writer.ReplyError(err.Error())
			return
		}
		writer.ReplySuccess(nil)
	}))

	return nil
}
