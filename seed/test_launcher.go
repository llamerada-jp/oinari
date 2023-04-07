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
package seed

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type testLauncher struct {
}

func NewTestLauncher() *testLauncher {
	return &testLauncher{}
}

func (r *testLauncher) Launch() error {
	fmt.Println("Run test")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	finCh := make(chan any, 1)
	hasFail := false

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				if string(arg.Value) == "\"FINISH\"" {
					finCh <- true
				}
			}

		case *runtime.EventExceptionThrown:
			// Since ts.URL uses a random port, replace it.
			s := ev.ExceptionDetails.Error()
			fmt.Printf("* %s\n", s)
		}
	})

	if err := chromedp.Run(ctx, chromedp.Navigate("http://localhost:8080/test.html")); err != nil {
		return err
	}

	<-finCh
	if hasFail {
		return fmt.Errorf("test failed")
	}

	return nil
}
