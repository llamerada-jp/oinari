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
package core

import (
	"log"
	"sync"

	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type Manager struct {
	mtx sync.Mutex
	cl  crosslink.Crosslink
	// key: containerID
	drivers map[string]CoreDriver
}

func NewCoreDriverManager(cl crosslink.Crosslink) *Manager {
	return &Manager{
		cl:      cl,
		drivers: make(map[string]CoreDriver),
	}
}

func (m *Manager) NewCoreDriver(containerID string, runtime []string) CoreDriver {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.drivers[containerID]; ok {
		log.Fatal("driver already exists")
	}

	var driver CoreDriver
L:
	for _, r := range runtime {
		switch r {
		case "core:dev1":
			driver = NewCoreAPIDriver(m.cl, containerID)
			break L
		}
	}

	if driver == nil {
		driver = NewNullAPIDriver()
	}

	m.drivers[containerID] = driver
	return driver
}

func (m *Manager) DestroyDriver(containerID string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.drivers, containerID)
}

func (m *Manager) GetDriver(containerID string) CoreDriver {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.drivers[containerID]
}
