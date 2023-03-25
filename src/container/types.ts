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

export const CrosslinkPath: string = "manager";

// ReadyRequest is a message to inform the parent that the worker has started
export interface ReadyRequest {
  // nothing
};

// ReadyReply informs the started worker of the program to run
export interface ReadyResponse {
  name: string
  // container_id: string
  // pod_sandbox_id: string
  image: ArrayBuffer
  // runtime: string
  args: string[]
  envs: Record<string, string>
};

// TermRequest is a message to inform the worker that let the program to terminate
export interface TermRequest {
  // nothing
}

export interface TermResponse {
  // nothing
}

// FinishedResponse is a message to inform the parent that the program has finished
export interface FinishedRequest {
  code: number
}

export interface FinishedResponse {
  // nothing
}