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
package misc

import (
	"log"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type dummyColonioValue struct {
	data []byte
}

var _ colonio.Value = &dummyColonioValue{}

type dummyColonio struct {
	kvs       map[string]*dummyColonioValue
	positionX float64
	positionY float64
}

var _ colonio.Colonio = &dummyColonio{}

// mimic Colonio for testings
func NewDummyColonio() *dummyColonio {
	return &dummyColonio{
		kvs: make(map[string]*dummyColonioValue),
	}
}

func (impl *dummyColonio) Connect(url, token string) error {
	log.Fatal("dummyColonio::Connect is not implemented")
	return nil
}

func (impl *dummyColonio) Disconnect() error {
	log.Fatal("dummyColonio::Disconnect is not implemented")
	return nil
}

func (impl *dummyColonio) IsConnected() bool {
	log.Fatal("dummyColonio::IsConnected is not implemented")
	return false
}

func (impl *dummyColonio) GetLocalNid() string {
	log.Fatal("dummyColonio::GetLocalNid is not implemented")
	return ""
}

func (impl *dummyColonio) SetPosition(x, y float64) (float64, float64, error) {
	impl.positionX = x
	impl.positionY = y
	return x, y, nil
}

func (impl *dummyColonio) GetLastPosition() (float64, float64) {
	return impl.positionX, impl.positionY
}

func (impl *dummyColonio) MessagingPost(dst, name string, val interface{}, opt uint32) (colonio.Value, error) {
	log.Fatal("dummyColonio::MessagingPost is not implemented")
	return nil, nil
}

func (impl *dummyColonio) MessagingSetHandler(name string, handler func(*colonio.MessagingRequest, colonio.MessagingResponseWriter)) {
	log.Fatal("dummyColonio::MessagingSetHandler is not implemented")
}

func (impl *dummyColonio) MessagingUnsetHandler(name string) {
	log.Fatal("dummyColonio::MessagingUnsetHandler is not implemented")
}

func (impl *dummyColonio) KvsGetLocalData() colonio.KvsLocalData {
	log.Fatal("dummyColonio::KvsGetLocalData is not implemented")
	return nil
}

func (impl *dummyColonio) KvsGet(key string) (colonio.Value, error) {
	v, ok := impl.kvs[key]
	if !ok {
		return nil, colonio.ErrKvsNotFound
	}
	return v, nil
}

func (impl *dummyColonio) KvsSet(key string, val interface{}, opt uint32) error {
	if opt&colonio.KvsProhibitOverwrite != 0 {
		if _, ok := impl.kvs[key]; ok {
			return colonio.ErrKvsProhibitOverwrite
		}
	}

	if val == nil {
		impl.kvs[key] = &dummyColonioValue{}
		return nil
	}

	switch v := val.(type) {
	case []byte:
		impl.kvs[key] = &dummyColonioValue{
			data: v,
		}
		return nil

	default:
		log.Fatal("dummyColonio::KvsSet is not support other than nil and []byte types")
		return nil
	}
}

func (impl *dummyColonio) SpreadPost(x, y, r float64, name string, message interface{}, opt uint32) error {
	log.Fatal("dummyColonio::SpreadPost is not implemented")
	return nil
}

func (impl *dummyColonio) SpreadSetHandler(name string, handler func(*colonio.SpreadRequest)) {
	log.Fatal("dummyColonio::SpreadSetHandler is not implemented")
}

func (impl *dummyColonio) SpreadUnsetHandler(name string) {
	log.Fatal("dummyColonio::SpreadUnsetHandler is not implemented")
}

func (impl *dummyColonioValue) IsNil() bool {
	return impl.data == nil
}

func (impl *dummyColonioValue) IsBool() bool {
	return false
}

func (impl *dummyColonioValue) IsInt() bool {
	return false
}

func (impl *dummyColonioValue) IsDouble() bool {
	return false
}

func (impl *dummyColonioValue) IsString() bool {
	return false
}

func (impl *dummyColonioValue) IsBinary() bool {
	return impl.data != nil
}

func (impl *dummyColonioValue) Set(val interface{}) error {
	if val == nil {
		impl.data = nil
	}

	switch v := val.(type) {
	case []byte:
		impl.data = v
		return nil

	default:
		log.Fatal("dummyColonioValue::Set is not support other than nil and []byte types")
		return nil
	}
}

func (impl *dummyColonioValue) GetBool() (bool, error) {
	return false, colonio.ErrUndefined
}

func (impl *dummyColonioValue) GetInt() (int64, error) {
	return 0, colonio.ErrUndefined
}

func (impl *dummyColonioValue) GetDouble() (float64, error) {
	return 0, colonio.ErrUndefined
}

func (impl *dummyColonioValue) GetString() (string, error) {
	return "", colonio.ErrUndefined
}

func (impl *dummyColonioValue) GetBinary() ([]byte, error) {
	if impl.data == nil {
		return nil, colonio.ErrUndefined
	}
	return impl.data, nil
}
