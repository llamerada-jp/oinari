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
package account

import (
	"context"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/stretchr/testify/suite"
)

const ACCOUNT = "cat"
const NODE_ID = "012345678901234567890123456789ab"

type AccountTest struct {
	suite.Suite
	col colonio.Colonio
	kvs KvsDriver
	mgr Manager
}

func NewAccountTest(ctx context.Context, col colonio.Colonio) *AccountTest {
	kvs := NewKvsDriver(col)

	return &AccountTest{
		col: col,
		kvs: kvs,
		mgr: NewManager(ctx, ACCOUNT, NODE_ID, kvs),
	}
}

func (at *AccountTest) TestGetAccountName() {
	at.Equal(ACCOUNT, at.mgr.GetAccountName())
}

func (at *AccountTest) TestRefresh() {

}
