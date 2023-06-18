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
	_ "embed"
	"encoding/json"
	"flag"
	"log"

	"github.com/llamerada-jp/colonio/go/service"
	"github.com/llamerada-jp/oinari/seed"
	"github.com/spf13/cobra"
)

var (
	//go:embed oinari.json
	configData []byte
	test       bool
)

var cmd = &cobra.Command{
	Use: "seed",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.CommandLine.Parse([]string{})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// setup colonio service
		config := &service.Config{}
		if err := json.Unmarshal(configData, config); err != nil {
			return err
		}

		sv, err := service.NewService(".", config, nil)
		if err != nil {
			return err
		}

		// publish debug path
		seed.InitDebug(sv.Mux)

		if test {
			// Start HTTP(S) service.
			go func() {
				if err := sv.Run(ctx); err != nil {
					log.Fatal(err)
				}
			}()

			// Run test
			launcher := seed.NewTestLauncher()
			return launcher.Launch()

		} else {
			return sv.Run(ctx)
		}
	},
}

func init() {
	flags := cmd.PersistentFlags()

	flags.BoolVarP(&test, "test", "t", false, "run test using headless browser")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
