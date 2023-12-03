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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/llamerada-jp/colonio/go/service"
	"github.com/llamerada-jp/oinari/seed"
	"github.com/spf13/cobra"
)

var (
	//go:embed commit_hash.txt
	commitHash []byte

	configPath string

	// run mode
	debug         bool
	test          bool
	withoutSignin bool
)

type oinariConfig struct {
	service.Config

	TemplateRoot    string `json:"templateRoot"`
	SecretParameter string `json:"secretParameter"`
}

var cmd = &cobra.Command{
	Use: "seed",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.CommandLine.Parse([]string{})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		config, path, err := readConfigFile()
		if err != nil {
			return err
		}

		secret, err := readSecret(filepath.Join(path, config.SecretParameter))
		if err != nil {
			return err
		}

		// setup colonio service
		sv, err := service.NewService(path, &config.Config, nil)
		if err != nil {
			return err
		}

		// setup seed handler
		seedInfo := &seed.SeedInfo{
			CommitHash: string(commitHash),
			Utime:      time.Now().Format(time.RFC3339),
		}
		if err := seed.Init(sv.Mux, secret, config.TemplateRoot, withoutSignin, seedInfo); err != nil {
			return err
		}

		// publish debug path
		if debug {
			if err := seed.InitDebug(sv.Mux); err != nil {
				return err
			}
		}

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

func readConfigFile() (*oinariConfig, string, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %w", err)
	}

	config := &oinariConfig{}
	if err := json.Unmarshal(raw, config); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %w", err)
	}

	path, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return nil, "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return config, path, nil
}

func readSecret(file string) (map[string]string, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	secret := map[string]string{}
	if err := json.Unmarshal(raw, &secret); err != nil {
		return nil, fmt.Errorf("failed to parse secret file: %w", err)
	}

	return secret, nil
}

func init() {
	flags := cmd.PersistentFlags()

	flags.StringVarP(&configPath, "config", "c", "./oinari.json", "config file path")
	flags.BoolVarP(&debug, "debug", "d", false, "run debug mode")
	flags.BoolVarP(&test, "test", "t", false, "run test using headless browser")
	flags.BoolVarP(&withoutSignin, "without-signin", "w", false, "run without signin")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
