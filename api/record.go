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

import "fmt"

type Record struct {
	Meta *ObjectMeta `json:"meta"`
	Data *RecordData `json:"data"`
}

type RecordEntry struct {
	Timestamp string `json:"timestamp"`
	Record    []byte
}

type RecordData struct {
	// key: container name
	Entries map[string]RecordEntry `json:"entries"`
}

func (record *Record) Validate() error {
	if err := record.Meta.Validate(ResourceTypeRecord); err != nil {
		return fmt.Errorf("invalid meta field: %s", err)
	}

	if err := ValidatePodUuid(record.Meta.Uuid); err != nil {
		return fmt.Errorf("invalid uuid field: %s", err)
	}

	if err := record.Data.Validate(); err != nil {
		return err
	}

	return nil
}

func (data *RecordData) Validate() error {
	if data.Entries == nil {
		return fmt.Errorf("entries field should not nil")
	}

	for _, entry := range data.Entries {
		if err := ValidateTimestamp(entry.Timestamp); err != nil {
			return fmt.Errorf("invalid timestamp: %s", err)
		}
	}

	return nil
}
