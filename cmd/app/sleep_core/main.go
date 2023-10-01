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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/llamerada-jp/oinari/lib/oinari"
)

type sleep struct {
	mtx         sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	durationSec uint64
	// for marshal
	PassedSec uint64 `json:"passedSec"`
}

func newSleep(ctx context.Context, durationSec uint64) *sleep {
	c, cancel := context.WithCancel(ctx)

	return &sleep{
		ctx:         c,
		cancel:      cancel,
		durationSec: durationSec,
	}
}

func (s *sleep) Setup(isInitialize bool, record []byte) error {
	if isInitialize || record == nil {
		fmt.Println("😪 start sleeping")
		s.PassedSec = 0

	} else {
		err := json.Unmarshal(record, s)
		if err != nil {
			return err
		}
		fmt.Printf("😪 continue to sleep for %d sec\n", s.PassedSec)
	}

	return s.start()
}

func (s *sleep) Marshal() ([]byte, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return json.Marshal(s)
}

func (s *sleep) Teardown(isFinalize bool) ([]byte, error) {
	s.cancel()
	// TODO: If use mutex, the container cannot stop. Is the limitation of WebAssembly thread?
	// s.mtx.Lock()
	// defer s.mtx.Unlock()

	if isFinalize {
		fmt.Println("😪 finish sleeping by interrupt")
	} else {
		fmt.Printf("😪 pause sleeping, %d sec passed\n", s.PassedSec)
	}

	return s.Marshal()
}

func (s *sleep) start() error {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			default:
				time.Sleep(1 * time.Second)
				s.mtx.Lock()
				s.PassedSec++
				if s.durationSec > 0 && s.PassedSec >= s.durationSec {
					fmt.Println("😪 finish sleeping by timeout")
					os.Exit(0)
				}
				s.mtx.Unlock()
			}
		}
	}()

	return nil
}

func showHelp(err error) {
	if err != nil {
		fmt.Printf("😪 %s\n", err.Error())
	}
	fmt.Printf("😪 usage: %s [duration]\n  duration: duration to sleep[sec], immediate wake up when 0, never wake up when negative value\n", os.Args[0])
}

func main() {
	var durationSec int64
	var err error

	if len(os.Args) == 2 {
		durationSec, err = strconv.ParseInt(os.Args[1], 10, 32)
		if err != nil {
			showHelp(err)
		}

	} else if len(os.Args) != 1 {
		showHelp(nil)
	}

	// wake up immediate if duration is 0
	if durationSec == 0 {
		os.Exit(0)
	}

	sleep := newSleep(context.Background(), uint64(durationSec))

	mgr := oinari.NewManager()
	err = mgr.Run(sleep)
	if err != nil {
		fmt.Println("😪", err)
		os.Exit(1)
	}

	os.Exit(0)
}
