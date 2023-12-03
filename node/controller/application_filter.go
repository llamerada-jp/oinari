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
package controller

import (
	"strings"

	coreAPI "github.com/llamerada-jp/oinari/api/core"
)

type ApplicationFilter interface {
	SetAccount(account string)
	SetFilter(filter string)
	SetSamplePrefix(prefix string)
	IsAllowed(pod *coreAPI.Pod) bool
}

type applicationFilterImpl struct {
	filter       string
	samplePrefix string
	account      string
}

func NewApplicationFilter() ApplicationFilter {
	return &applicationFilterImpl{
		filter: "samples",
	}
}

func (impl *applicationFilterImpl) SetAccount(account string) {
	impl.account = account
}

func (impl *applicationFilterImpl) SetFilter(filter string) {
	impl.filter = filter
}

func (impl *applicationFilterImpl) SetSamplePrefix(prefix string) {
	impl.samplePrefix = prefix
}

func (impl *applicationFilterImpl) IsAllowed(pod *coreAPI.Pod) bool {
	if impl.filter == "any" {
		return true
	}

	if impl.filter == "samples" || impl.filter == "samplesAndMyself" {
		isSample := true
		for _, container := range pod.Spec.Containers {
			if !strings.HasPrefix(container.Image, impl.samplePrefix) {
				isSample = false
				break
			}
		}
		if isSample {
			return true
		}
	}

	if impl.filter == "myself" || impl.filter == "samplesAndMyself" {
		if pod.Meta.Owner == impl.account {
			return true
		}
	}

	return false
}
