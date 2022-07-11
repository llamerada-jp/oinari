package core

import (
	"context"
	"encoding/json"
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type PodImpl struct {
	uuid string
}

func NewPod(uuid string) *PodImpl {
	return &PodImpl{
		uuid: uuid,
	}
}

// return false if the pod can be deleted.
func (p *PodImpl) Update(ctx context.Context, colonio colonio.Colonio, resources colonio.Map) (bool, error) {
	log.Println("update", p.uuid)
	key := string(ResourceTypePod) + "/" + p.uuid
	// TODO colonio has a bug ? that the `Get`` method is blocked when it used in cb?
	v, err := resources.Get(key)
	if err != nil {
		return false, err
	}

	s, err := v.GetString()
	if err != nil {
		return false, err
	}

	var pod Pod
	err = json.Unmarshal([]byte(s), &pod)
	if err != nil {
		return false, err
	}

	localNid := colonio.GetLocalNid()
	if pod.Status.RunningNode == localNid {
		log.Println("fixme!")
		return true, nil
	}

	if pod.Status.RunningNode == "" {
		log.Println("fixme!")
		return true, nil
	}

	return false, nil
}

func (p *PodImpl) HasLock() bool {
	// TODO return lock status of lease
	return true
}
