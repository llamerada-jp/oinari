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

type nullAPIDriverImpl struct {
}

func NewNullAPIDriver() CoreDriver {
	return &nullAPIDriverImpl{}
}

func (driver *nullAPIDriverImpl) DriverName() string {
	return ""
}

func (driver *nullAPIDriverImpl) Setup(firstInPod bool) error {
	return nil
}

func (driver *nullAPIDriverImpl) Dump() ([]byte, error) {
	return nil, nil
}

func (driver *nullAPIDriverImpl) Restore(dumpData []byte) error {
	return nil
}

func (driver *nullAPIDriverImpl) Teardown(lastInPod bool) error {
	return nil
}
