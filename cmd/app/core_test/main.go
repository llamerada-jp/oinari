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

package coretest

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/llamerada-jp/oinari/lib/oinari"
)

type coreTest struct {
	CountSetup    int `json:"setup"`
	CountDump     int `json:"dump"`
	CountRestore  int `json:"restore"`
	CountTeardown int `json:"teardown"`
}

type logEntry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Param any    `json:"param"`
}

func (ct *coreTest) Setup(firstInPod bool) error {
	ct.CountSetup++
	return ct.outputLog("Setup", ct.CountSetup, firstInPod)
}

func (ct *coreTest) Dump() ([]byte, error) {
	ct.CountDump++
	err := ct.outputLog("Dump", ct.CountDump, nil)
	if err != nil {
		return nil, err
	}

	return json.Marshal(*ct)
}

func (ct *coreTest) Restore(data []byte) error {
	ct.CountRestore++
	err := ct.outputLog("Restore", ct.CountRestore, nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ct)
}

func (ct *coreTest) Teardown(lastInPod bool) error {
	ct.CountTeardown++
	return ct.outputLog("Teardown", ct.CountTeardown, lastInPod)
}

func (ct *coreTest) outputLog(name string, count int, param any) error {
	entry := logEntry{
		Name:  name,
		Count: count,
		Param: param,
	}

	entryJS, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(oinari.Writer, string(entryJS))
	return err
}

func Main() {
	ct := &coreTest{}
	err := oinari.Run(ct)
	if err != nil {
		log.Fatal(err)
	}
}
