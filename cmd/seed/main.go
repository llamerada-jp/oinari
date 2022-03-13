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
	"time"

	"github.com/llamerada-jp/colonio/go/seed"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/acme/autocert"
)

var (
	//go:embed oinari.json
	configData   []byte
	documentRoot string
	port         uint16
	fqdn         string
	debugMode    bool
)

var cmd = &cobra.Command{
	Use: "seed",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.CommandLine.Parse([]string{})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		mux := http.NewServeMux()

		if debugMode {
			enableDebugMode()

			utime := time.Now().Format(time.RFC3339)
			mux.HandleFunc("/utime", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(utime))
			})
		}

		// Set HTTP headers to enable SharedArrayBuffer for WebAssembly Threads
		headers := make(map[string]string)
		headers["Cross-Origin-Embedder-Policy"] = "require-corp"
		headers["Cross-Origin-Opener-Policy"] = "same-origin"

		// Publish static documents using HTTP(S).
		mux.Handle("/", &headerWrapper{
			handler: http.FileServer(http.Dir(documentRoot)),
			headers: headers,
		})

		// Publish colonio-seed
		seedConfig := &seed.Config{}
		if err := json.Unmarshal(configData, seedConfig); err != nil {
			return err
		}
		seed, err := seed.NewSeed(seedConfig)
		if err != nil {
			return err
		}
		mux.Handle("/seed", seed)
		if err := seed.Start(); err != nil {
			return err
		}

		// Start HTTP(S) service.
		if len(fqdn) != 0 {
			return http.Serve(autocert.NewListener(fqdn), mux)
		} else {
			return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		}
	},
}

type headerWrapper struct {
	handler http.Handler
	headers map[string]string
}

func (h *headerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headerWriter := w.Header()
	for k, v := range h.headers {
		headerWriter.Add(k, v)
	}
	h.handler.ServeHTTP(w, r)
}

func init() {
	flags := cmd.PersistentFlags()

	flags.StringVarP(&documentRoot, "root", "r", "./dist", "path to document root")
	flags.Uint16VarP(&port, "port", "p", 80, "port number for HTTP service, and it will be ignored if https option is selected")
	flags.StringVarP(&fqdn, "https", "s", "", "enable HTTPS service and set FQDN, and certificate is configured using Let's encrypt automatically")
	flags.BoolVarP(&debugMode, "debug", "d", false, "enable debug mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
