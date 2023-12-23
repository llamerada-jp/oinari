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
package three

type CreateObjectRequest struct {
	Name string      `json:"name"`
	Spec *ObjectSpec `json:"spec"`
}

type CreateObjectResponse struct {
	UUID string `json:"uuid"`
}

// This method is foolish and inefficient and needs to be improved.
type UpdateObjectRequest struct {
	UUID string      `json:"uuid"`
	Spec *ObjectSpec `json:"spec"`
}

type UpdateObjectResponse struct {
	// empty
}

type GetObjectRequest struct {
	UUID string `json:"uuid"`
}

type GetObjectResponse struct {
	Object *Object `json:"object"`
}

type DeleteObjectRequest struct {
	UUID string `json:"uuid"`
}

type DeleteObjectResponse struct {
	// empty
}
