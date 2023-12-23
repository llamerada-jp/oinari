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
	"log"
	"os"
	"strconv"
	"time"
)

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

	// never wake up if duration is negative value
	if durationSec < 0 {
		for {
			time.Sleep(time.Hour)
		}
	}

	// sleep for specified duration
	time.Sleep(time.Second * time.Duration(durationSec))
	os.Exit(0)
}
