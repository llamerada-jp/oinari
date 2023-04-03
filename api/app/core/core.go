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

package app

// types to pass from application to node manager
type ReadyRequest struct {
	// empty
}

type ReadyResponse struct {
	// empty
}

type OutputRequest struct {
	Payload []byte `json:"payload"`
}

type OutputResponse struct {
	Length int `json:"length"`
}

// types to pass from node manager to application
type SetupRequest struct {
	FirstInPod bool `json:"firstInPod"`
}

type SetupResponse struct {
	// empty
}

type DumpRequest struct {
	// empty
}

type DumpResponse struct {
	DumpData []byte `json:"dumpData"`
}

type RestoreRequest struct {
	DumpData []byte `json:"dumpData"`
}

type RestoreResponse struct {
	// empty
}

type TeardownRequest struct {
	LastInPod bool `json:"lastInPod"`
}

type TeardownResponse struct {
	// empty
}
