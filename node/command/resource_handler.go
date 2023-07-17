package command

import (
	"github.com/llamerada-jp/oinari/api"
	"github.com/llamerada-jp/oinari/lib/crosslink"
	"github.com/llamerada-jp/oinari/node/resource/account"
	"github.com/llamerada-jp/oinari/node/resource/node"
	"github.com/llamerada-jp/oinari/node/resource/pod"
)

type createPodRequest struct {
	Name string       `json:"name"`
	Spec *api.PodSpec `json:"spec"`
}

type createPodResponse struct {
	Digest *pod.ApplicationDigest `json:"digest"`
}

type listPodResponse struct {
	Digests []pod.ApplicationDigest `json:"digests"`
}

type deletePodRequest struct {
	Uuid string `json:"uuid"`
}

func InitResourceHandler(rootMpx crosslink.MultiPlexer, accountMgr account.Manager, nodeMgr node.Manager, podMgr pod.Manager) {
	mpx := crosslink.NewMultiPlexer()
	rootMpx.SetHandler("resource", mpx)

	mpx.SetHandler("createPod", crosslink.NewFuncHandler(
		func(request *createPodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			digest, err := podMgr.Create(request.Name, accountMgr.GetAccountName(), nodeMgr.GetNid(), request.Spec)
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			writer.ReplySuccess(createPodResponse{
				Digest: digest,
			})
		}))

	mpx.SetHandler("listPod", crosslink.NewFuncHandler(
		func(request *interface{}, tags map[string]string, writer crosslink.ResponseWriter) {
			// TODO
			writer.ReplySuccess(nil)
		}))

	mpx.SetHandler("deletePod", crosslink.NewFuncHandler(
		func(param *deletePodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			// TODO
			writer.ReplySuccess(nil)
		}))
}
