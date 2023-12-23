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
	"encoding/json"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type testLauncher struct {
}

func NewTestLauncher() *testLauncher {
	return &testLauncher{}
}

func (r *testLauncher) Launch() error {
	log.Println("Run test")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"),
	)
	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel1()
	ctx, cancel2 := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(log.Printf))
	defer cancel2()

	resultCh := make(chan bool, 1)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				if arg.Type != runtime.TypeString {
					continue
				}
				var line string
				if err := json.Unmarshal(arg.Value, &line); err != nil {
					log.Println("failed to decode: ", line)
					continue
				}
				log.Println("(browser)", line)

				switch line {
				case "SUCCESS":
					resultCh <- true

				case "FAIL":
					resultCh <- false
				}
			}

		case *runtime.EventExceptionThrown:
			// Since ts.URL uses a random port, replace it.
			s := ev.ExceptionDetails.Error()
			fmt.Printf("* %s\n", s)
		}
	})

	if err := chromedp.Run(ctx, chromedp.Navigate("https://localhost:8080/test.html")); err != nil {
		return err
	}

	result := <-resultCh
	if !result {
		return fmt.Errorf("test failed")
	}

	return nil
}
