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
	"log"
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
	PassedSec uint64
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
	if isInitialize {
		fmt.Println("start sleeping")
		s.PassedSec = 0

	} else {
		fmt.Printf("continue to sleep for %d sec", s.PassedSec)
		err := json.Unmarshal(record, s)
		if err != nil {
			return err
		}
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

	if isFinalize {
		fmt.Println("finish sleeping by interrupt")
	} else {
		fmt.Println("pause sleeping")
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
					fmt.Println("finish sleeping by timeout")
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
		log.Println(err)
	}
	log.Fatalf("usage: %s [duration]\n  duration: duration to sleep[sec], immediate wake up when 0, never wake up when negative value", os.Args[0])
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

	err = oinari.Run(sleep)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
