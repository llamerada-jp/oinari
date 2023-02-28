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
	"encoding/json"
	htmlTemplate "html/template"
	"os"
	"strings"
	textTempalte "text/template"

	"github.com/spf13/cobra"
)

var (
	inputFile  string
	valuesFile string
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "generate output file using template and parameter files.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		valuesJson, err := os.ReadFile(valuesFile)
		if err != nil {
			return err
		}

		values := make(map[string]interface{})
		if err = json.Unmarshal(valuesJson, &values); err != nil {
			return err
		}

		if strings.HasSuffix(inputFile, ".html") || strings.HasSuffix(inputFile, ".htm") {
			tpl, err := htmlTemplate.ParseFiles(inputFile)
			if err != nil {
				return err
			}

			if err = tpl.Execute(os.Stdout, values); err != nil {
				return err
			}

		} else {
			tpl, err := textTempalte.ParseFiles(inputFile)
			if err != nil {
				return err
			}

			if err = tpl.Execute(os.Stdout, values); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	flags := templateCmd.PersistentFlags()
	flags.StringVarP(&inputFile, "in", "i", "input.txt", "A template file that uses golang's template format.")
	flags.StringVarP(&valuesFile, "values", "v", "values.json", "A JSON file with the values to be embedded in the template.")
	rootCmd.AddCommand(templateCmd)
}
