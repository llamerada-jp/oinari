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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNodeId(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(ValidateNodeId("01234567890123456789012345abcdef"))

	for _, nid := range []string{
		"",
		"01234567890123456789012345abcde",
		"g1234567890123456789012345abcdef",
		"01234567890123456789012345ABCDEF",
	} {
		assert.Error(ValidateNodeId(nid), nid)
	}
}

func TestValidateTimestamp(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(ValidateTimestamp("2021-04-09T14:00:40+09:00"))

	for _, timestamp := range []string{
		"",
		"2021/04/09T14:00:40+09:00",
	} {
		assert.Error(ValidateTimestamp(timestamp))
	}
}
