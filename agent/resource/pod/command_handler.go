package pod

import (
	"github.com/llamerada-jp/oinari/agent/crosslink"
	"github.com/llamerada-jp/oinari/api"
)

type applicationDigest struct {
	Name        string `json:"name"`
	Uuid        string `json:"uuid"`
	RunningNode string `json:"runningNode"`
	Owner       string `json:"owner"`
}

type runRequest struct {
	Name string       `json:"name"`
	Spec *api.PodSpec `json:"spec"`
}

type runResponse struct {
	Digest *applicationDigest `json:"digest"`
}

type listResponse struct {
	Digests []applicationDigest `json:"digests"`
}

type terminateRequest struct {
	Uuid string `json:"uuid"`
}

func InitCommandHandler(podMgr Manager, rootHandler crosslink.MultiPlexer) error {
	mpx := crosslink.NewMultiPlexer()
	rootHandler.SetHandler("pod_manager", mpx)

	mpx.SetHandler("run", crosslink.NewFuncHandler(
		func(request *runRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			digest, err := podMgr.run(request.Name, request.Spec)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(runResponse{
				Digest: digest,
			})
		}))

	mpx.SetHandler("list", crosslink.NewFuncHandler(
		func(request *interface{}, tags map[string]string, writer crosslink.ResponseWriter) {
			list, err := podMgr.list()
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(listResponse{
				Digests: list,
			})
		}))

	mpx.SetHandler("terminate", crosslink.NewFuncHandler(
		func(param *terminateRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			err := podMgr.terminate(param.Uuid)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(nil)
		}))

	return nil
}
