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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or usied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package misc

import "sync"

type UniqueSet struct {
	mtx  *sync.Mutex
	cond *sync.Cond
	keys map[string]bool
}

func NewUniqueSet() *UniqueSet {
	mtx := &sync.Mutex{}
	cond := sync.NewCond(mtx)

	return &UniqueSet{
		mtx:  mtx,
		cond: cond,
		keys: make(map[string]bool),
	}
}

func (us *UniqueSet) Insert(key string) {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	for us.keys[key] {
		us.cond.Wait()
	}
	us.keys[key] = true
}

func (us *UniqueSet) Remove(key string) {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	delete(us.keys, key)
	us.cond.Broadcast()
}
