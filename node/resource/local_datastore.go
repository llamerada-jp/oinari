package resource

import (
	"fmt"
	"strings"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type LocalResource struct {
	ResourceType api.ResourceType
	RecordRaw    []byte
}

type LocalDatastore interface {
	GetResources() ([]LocalResource, error)
}

type localDatastore struct {
	col colonio.Colonio
}

func NewLocalDatastore(col colonio.Colonio) LocalDatastore {
	return &localDatastore{
		col: col,
	}
}

func (ld *localDatastore) GetResources() ([]LocalResource, error) {
	resources := make([]LocalResource, 0)

	// to avoid dead-lock of ForeachLocalValue, don't call colonio's method in the callback func
	localData := ld.col.KvsGetLocalData()
	defer localData.Free()

	for _, key := range localData.GetKeys() {
		v, err := localData.GetValue(key)
		if err != nil {
			return nil, err
		}
		raw, err := v.GetBinary()
		if err != nil {
			return nil, err
		}
		resourceEntry := strings.Split(key, "/")
		if len(resourceEntry) != 2 {
			return nil, fmt.Errorf("local value key is not supported format:%s", key)
		}
		resources = append(resources, LocalResource{
			ResourceType: api.ResourceType(resourceEntry[0]),
			RecordRaw:    raw,
		})
	}

	return resources, nil
}
