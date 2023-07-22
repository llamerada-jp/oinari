package command

import (
	"log"

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
			res := listPodResponse{
				Digests: make([]pod.ApplicationDigest, 0),
			}

			uuids := make(map[string]any, 0)

			// get pod uuids bound for the account
			podState, err := accountMgr.GetAccountPodState()
			if err != nil {
				writer.ReplyError(err.Error())
				return
			}
			for uuid := range podState {
				uuids[uuid] = true
			}

			// get pod uuids running on local node
			for _, uuid := range podMgr.GetLocalPodUUIDs() {
				uuids[uuid] = true
			}

			// make pod digest
			for uuid := range uuids {
				p, err := podMgr.GetPodData(uuid)
				if err != nil {
					log.Printf("error on get pod info: %s", err.Error())
					continue
				}
				res.Digests = append(res.Digests, pod.ApplicationDigest{
					Name:        p.Meta.Name,
					Uuid:        uuid,
					RunningNode: p.Status.RunningNode,
					Owner:       p.Meta.Owner,
					Phase:       string(p.Status.Phase),
				})
			}

			writer.ReplySuccess(res)
		}))

	mpx.SetHandler("deletePod", crosslink.NewFuncHandler(
		func(param *deletePodRequest, tags map[string]string, writer crosslink.ResponseWriter) {
			// TODO
			writer.ReplySuccess(nil)
		}))
}
