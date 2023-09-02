/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package mock

import (
	"log"
	"sync"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type colonioValue struct {
	data []byte
}

var _ colonio.Value = &colonioValue{}

type Colonio struct {
	mutex     sync.Mutex
	kvs       map[string]*colonioValue
	positionX float64
	positionY float64
}

var _ colonio.Colonio = &Colonio{}

// mimic Colonio for testings
func NewColonioMock() *Colonio {
	return &Colonio{
		kvs: make(map[string]*colonioValue),
	}
}

func (impl *Colonio) DeleteKVSAll() {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()

	impl.kvs = make(map[string]*colonioValue)
}

func (impl *Colonio) Connect(url, token string) error {
	log.Fatal("ColonioMock::Connect is not implemented")
	return nil
}

func (impl *Colonio) Disconnect() error {
	log.Fatal("ColonioMock::Disconnect is not implemented")
	return nil
}

func (impl *Colonio) IsConnected() bool {
	log.Fatal("ColonioMock::IsConnected is not implemented")
	return false
}

func (impl *Colonio) GetLocalNid() string {
	log.Fatal("ColonioMock::GetLocalNid is not implemented")
	return ""
}

func (impl *Colonio) SetPosition(x, y float64) (float64, float64, error) {
	impl.positionX = x
	impl.positionY = y
	return x, y, nil
}

func (impl *Colonio) GetLastPosition() (float64, float64) {
	return impl.positionX, impl.positionY
}

func (impl *Colonio) MessagingPost(dst, name string, val interface{}, opt uint32) (colonio.Value, error) {
	log.Fatal("ColonioMock::MessagingPost is not implemented")
	return nil, nil
}

func (impl *Colonio) MessagingSetHandler(name string, handler func(*colonio.MessagingRequest, colonio.MessagingResponseWriter)) {
	log.Fatal("ColonioMock::MessagingSetHandler is not implemented")
}

func (impl *Colonio) MessagingUnsetHandler(name string) {
	log.Fatal("ColonioMock::MessagingUnsetHandler is not implemented")
}

func (impl *Colonio) KvsGetLocalData() colonio.KvsLocalData {
	log.Fatal("ColonioMock::KvsGetLocalData is not implemented")
	return nil
}

func (impl *Colonio) KvsGet(key string) (colonio.Value, error) {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()

	v, ok := impl.kvs[key]
	if !ok {
		return nil, colonio.ErrKvsNotFound
	}
	return v, nil
}

func (impl *Colonio) KvsSet(key string, val interface{}, opt uint32) error {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()

	if opt&colonio.KvsProhibitOverwrite != 0 {
		if _, ok := impl.kvs[key]; ok {
			return colonio.ErrKvsProhibitOverwrite
		}
	}

	if val == nil {
		impl.kvs[key] = &colonioValue{}
		return nil
	}

	switch v := val.(type) {
	case []byte:
		impl.kvs[key] = &colonioValue{
			data: v,
		}
		return nil

	default:
		log.Fatal("ColonioMock::KvsSet is not support other than nil and []byte types")
		return nil
	}
}

func (impl *Colonio) SpreadPost(x, y, r float64, name string, message interface{}, opt uint32) error {
	log.Fatal("ColonioMock::SpreadPost is not implemented")
	return nil
}

func (impl *Colonio) SpreadSetHandler(name string, handler func(*colonio.SpreadRequest)) {
	log.Fatal("ColonioMock::SpreadSetHandler is not implemented")
}

func (impl *Colonio) SpreadUnsetHandler(name string) {
	log.Fatal("ColonioMock::SpreadUnsetHandler is not implemented")
}

func (impl *colonioValue) IsNil() bool {
	return impl.data == nil
}

func (impl *colonioValue) IsBool() bool {
	return false
}

func (impl *colonioValue) IsInt() bool {
	return false
}

func (impl *colonioValue) IsDouble() bool {
	return false
}

func (impl *colonioValue) IsString() bool {
	return false
}

func (impl *colonioValue) IsBinary() bool {
	return impl.data != nil
}

func (impl *colonioValue) Set(val interface{}) error {
	if val == nil {
		impl.data = nil
	}

	switch v := val.(type) {
	case []byte:
		impl.data = v
		return nil

	default:
		log.Fatal("colonioValueMock::Set is not support other than nil and []byte types")
		return nil
	}
}

func (impl *colonioValue) GetBool() (bool, error) {
	return false, colonio.ErrUndefined
}

func (impl *colonioValue) GetInt() (int64, error) {
	return 0, colonio.ErrUndefined
}

func (impl *colonioValue) GetDouble() (float64, error) {
	return 0, colonio.ErrUndefined
}

func (impl *colonioValue) GetString() (string, error) {
	return "", colonio.ErrUndefined
}

func (impl *colonioValue) GetBinary() ([]byte, error) {
	if impl.data == nil {
		return nil, colonio.ErrUndefined
	}
	return impl.data, nil
}
