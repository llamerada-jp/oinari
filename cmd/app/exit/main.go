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
)

func showHelp(err error) {
	if err != nil {
		log.Printf("`exit` program failed: %s", err.Error())
	}
	log.Fatalf("usage: %s [code]\n  code: exit code, default 0", os.Args[0])
}

func main() {
	var code int64
	var err error

	if len(os.Args) == 2 {
		code, err = strconv.ParseInt(os.Args[1], 10, 8)
		if err != nil {
			showHelp(err)
		}

	} else if len(os.Args) != 1 {
		showHelp(nil)

	}

	os.Exit(int(code))
}
