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
package api

import (
	"fmt"
	"regexp"
	"time"
)

// node id format is equal to colonio node ids
var NODE_NAME_EXPRESSION = regexp.MustCompile("^[0-9a-z]{32}$")

func ValidateNodeId(name string) error {
	if !NODE_NAME_EXPRESSION.Match([]byte(name)) {
		return fmt.Errorf("node id should be 128bit hex format")
	}
	return nil
}

func ValidateTimestamp(timestamp string) error {
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("timestamp should be ISO8601/RFC3339 format:%w", err)
	}
	return nil
}
