package pod

import (
	"encoding/json"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type KvsDriver interface {
	createPod(pod *api.Pod) error
	updatePod(pod *api.Pod) error
}

type kvsDriverImpl struct {
	col colonio.Colonio
}

func NewKvsDriver(col colonio.Colonio) KvsDriver {
	return &kvsDriverImpl{
		col: col,
	}
}

func (kvs *kvsDriverImpl) createPod(pod *api.Pod) error {
	if err := pod.Validate(true); err != nil {
		return err
	}

	key := string(api.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	return kvs.col.KvsSet(key, raw, colonio.KvsProhibitOverwrite)
}

func (kvs *kvsDriverImpl) updatePod(pod *api.Pod) error {
	if err := pod.Validate(true); err != nil {
		return err
	}

	key := string(api.ResourceTypePod) + "/" + pod.Meta.Uuid
	raw, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	return kvs.col.KvsSet(key, raw, 0)
}
