package resource

import (
	"context"
	"log"
	"time"

	"github.com/llamerada-jp/oinari/agent/resource/pod"
	"github.com/llamerada-jp/oinari/api"
)

type Manager interface {
	Start(ctx context.Context) error
}

type manager struct {
	ld     LocalDatastore
	podMgr pod.Manager
}

func NewManager(ld LocalDatastore, podMtr pod.Manager) Manager {
	return &manager{
		ld:     ld,
		podMgr: podMtr,
	}
}

func (mgr *manager) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			err := mgr.loop(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (mgr *manager) loop(ctx context.Context) error {
	resources, err := mgr.ld.GetResources()
	if err != nil {
		return err
	}

	for _, resource := range resources {
		switch resource.ResourceType {
		case api.ResourceTypePod:
			err = mgr.podMgr.DealLocalResource(resource.RecordRaw)
		}
		if err != nil {
			return err
		}
	}

	err = mgr.podMgr.Loop(ctx)
	if err != nil {
		return err
	}

	return nil
}
