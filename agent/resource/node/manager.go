package node

type Manager interface {
	GetNid() string
}

type ManagerImpl struct {
	localNid string
}

func NewManager(localNid string) *ManagerImpl {
	return &ManagerImpl{
		localNid: localNid,
	}
}

func (mgr *ManagerImpl) GetNid() string {
	return mgr.localNid
}
