package account

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/api"
)

type KvsDriver interface {
	setMeta(name string) error
	bindPod(pod *api.Pod) error
}

type kvsDriverImpl struct {
	name string
	col  colonio.Colonio
}

func NewKvsDriver(col colonio.Colonio) KvsDriver {
	return &kvsDriverImpl{
		col: col,
	}
}

func (kvs *kvsDriverImpl) setMeta(name string) error {
	kvs.name = name

	account, err := kvs.getOrCreate(name)
	if err != nil {
		return err
	}

	if account.Meta.Name != name {
		return fmt.Errorf("account uuid collision")
	}

	return nil
}

func (kvs *kvsDriverImpl) bindPod(pod *api.Pod) error {
	account, err := kvs.getOrCreate(kvs.name)
	if err != nil {
		return err
	}

	node, ok := account.Status.Pods[pod.Meta.Uuid]
	if ok && node == pod.Status.RunningNode {
		return nil
	}

	account.Status.Pods[pod.Meta.Uuid] = pod.Status.RunningNode
	return kvs.set(account)
}

func (kvs *kvsDriverImpl) getOrCreate(name string) (*api.Account, error) {
	key := kvs.getKey(name)
	val, err := kvs.col.KvsGet(key)
	if err == nil {
		raw, err := val.GetBinary()
		if err != nil {
			return nil, err
		}

		var account api.Account
		err = json.Unmarshal(raw, &account)
		if err != nil {
			return nil, err
		}

		return &account, nil
	}
	// TODO: check does account data exist?

	account := &api.Account{
		Meta: &api.ObjectMeta{
			Type:    api.ResourceTypeAccount,
			Name:    name,
			Account: name,
			Uuid:    kvs.getUUID(name),
		},
		Status: &api.AccountStatus{
			Pods: make(map[string]string),
		},
	}

	err = kvs.set(account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (kvs *kvsDriverImpl) set(account *api.Account) error {
	raw, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = kvs.col.KvsSet(kvs.getKey(account.Meta.Name), raw, 0)
	if err != nil {
		return err
	}

	return nil
}

// use sha256 hash as account's uuid
func (kvs *kvsDriverImpl) getKey(name string) string {
	return string(api.ResourceTypeAccount) + "/" + kvs.getUUID(name)
}

// use sha256 hash as account's uuid
func (kvs *kvsDriverImpl) getUUID(name string) string {
	hash := sha256.Sum256([]byte(kvs.name))
	return hex.EncodeToString(hash[:])
}
