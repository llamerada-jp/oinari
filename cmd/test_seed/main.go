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
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	colonioSeed "github.com/llamerada-jp/colonio/go/seed"
	"github.com/llamerada-jp/oinari/seed"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	//go:embed oinari.json
	configData   []byte
	documentRoot string
	port         uint16
	test         bool
)

var cmd = &cobra.Command{
	Use: "seed",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.CommandLine.Parse([]string{})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		mux := http.NewServeMux()

		seed.InitDebug(mux)

		// Publish colonio-seed
		seedConfig := &colonioSeed.Config{}
		if err := json.Unmarshal(configData, seedConfig); err != nil {
			return err
		}

		seed.InitSeed(mux, documentRoot, seedConfig)

		if test {
			// Start HTTP(S) service.
			go func() {
				if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
					log.Fatal(err)
				}
			}()

			// Run test
			launcher := seed.NewTestLauncher()
			if err := launcher.Launch(); err != nil {
				return err
			}

		} else {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
				log.Fatal(err)
			}
		}

		return nil
	},
}

func init() {
	flags := cmd.PersistentFlags()

	flags.StringVarP(&documentRoot, "root", "r", "./dist", "path to document root")
	flags.Uint16VarP(&port, "port", "p", 8080, "port number for HTTP service, and it will be ignored if https option is selected")
	flags.BoolVarP(&test, "test", "t", false, "run test using headless browser")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
