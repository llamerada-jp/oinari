package account

import (
	"github.com/llamerada-jp/oinari/api"
)

type Manager interface {
	GetAccountName() string
	Refresh() error
	BindPod(pod *api.Pod) error
}

type ManagerImpl struct {
	name string
	kvs  KvsDriver
}

func NewManager(account string, kvs KvsDriver) *ManagerImpl {
	return &ManagerImpl{
		name: account,
		kvs:  kvs,
	}
}

func (mgr *ManagerImpl) GetAccountName() string {
	return mgr.name
}

func (mgr *ManagerImpl) Refresh() error {
	return mgr.kvs.setMeta(mgr.name)
}

func (mgr *ManagerImpl) BindPod(pod *api.Pod) error {
	return mgr.kvs.bindPod(pod)
}
