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

import * as CL from "./crosslink";
import * as CT from "./container/types";

const crosslinkPath: string = "cri";
const podSandboxIDMax: number = Math.floor(Math.pow(2, 30));
const containerIDMax: number = Math.floor(Math.pow(2, 30));
const imageRefIDMax: number = Math.floor(Math.pow(2, 30));
const containerStopTimeout: number = 10 * 1000;

// key means PodSandboxID
let sandboxes: Map<string, SandboxImpl> = new Map();
// key means ContainerID
let containers: Map<string, ContainerImpl> = new Map();
// key means image url (not image ref id)
let images: Map<string, ImageImpl> = new Map();

function getTimestamp(): string {
  return new Date().toISOString();
}

enum ImageState {
  Created = 0,
  Downloading = 1,
  Downloaded = 2,
  Error = 3,
}

class ImageImpl {
  state: ImageState
  // image ref id
  id: string
  url: string
  runtime: string
  image: ArrayBuffer | undefined

  constructor(id: string, url: string) {
    this.state = ImageState.Created;
    this.id = id;
    this.url = url;
    this.runtime = "";
    this.pull();
  }

  pull() {
    console.assert(this.state !== ImageState.Downloading, "duplicate download");

    this.state = ImageState.Downloading;

    fetch(this.url).then((response: Response) => {
      if (!response.ok) {
        throw new Error(response.statusText);
      }
      return response.json();

    }).then((p) => {
      let param = p as {
        image: string
        runtime: string
      };

      // TODO check runtime
      this.runtime = param.runtime;

      return fetch(new URL(param.image, this.url).toString());

    }).then((response: Response) => {
      if (!response.ok || response.status !== 200) {
        throw new Error(response.statusText);
      }

      return response.arrayBuffer();

    }).then((buffer) => {
      if (buffer == null) {
        throw new Error("array buffer shouldn't be null");
      }
      this.state = ImageState.Downloaded;
      this.image = buffer;

    }).catch((reason) => {
      this.state = ImageState.Error;
      console.error(reason);
    });
  }
}

class ContainerImpl {
  id: string
  // PodSandboxID
  sandbox_id: string
  name: string
  image: ImageImpl
  worker: Worker | undefined
  link: CL.Crosslink | undefined
  args: string[]
  envs: Record<string, string>
  created_at: string
  started_at: string | undefined
  finished_at: string | undefined
  exit_code: number | undefined

  constructor(id: string, sandbox_id: string, name: string, image: ImageImpl, args: string[], envs: Record<string, string>) {
    this.id = id;
    this.sandbox_id = sandbox_id;
    this.name = name;
    this.image = image;
    this.args = args;
    this.envs = envs;
    this.created_at = getTimestamp();
  }

  getState(): ContainerState {
    if (this.finished_at != null) {
      return ContainerState.ContainerExited;
    }
    if (this.started_at != null) {
      return ContainerState.ContainerRunning;
    }
    return ContainerState.ContainerCreated;
  }

  start() {
    console.assert(this.worker == null);

    let rootMpx = new CL.MultiPlexer();
    this.worker = new Worker("container.js");
    this.link = new CL.Crosslink(new CL.WorkerImpl(this.worker), rootMpx);

    this._initHandler(rootMpx);
  }

  stop() {
    if (this.worker == null) {
      return;
    }

    this.link?.call(CT.CrosslinkPath + "/term", {});
    setTimeout(()=>{
      if (this.finished_at != null) {
        this.finished_at = getTimestamp();
        // 137 meaning 128 + sig kill(9)
        this.exit_code = 137;
      }

      this.cleanup();
    }, containerStopTimeout);
  }

  cleanup() {
    if (this.worker == null) {
      return;
    }

    this.worker.terminate();
    this.worker = undefined;
    this.link = undefined;
  }

  _initHandler(rootMpx: CL.MultiPlexer) {
    let mpx = new CL.MultiPlexer();
    rootMpx.setHandler(CT.CrosslinkPath, mpx);

    mpx.setObjHandlerFunc("ready", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      let res = this._onReady(data as CT.ReadyRequest);
      writer.replySuccess(res);
    });

    mpx.setObjHandlerFunc("finished", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      let res = this._onFinished(data as CT.FinishedRequest);
      writer.replySuccess(res);
    });
  }

  _onReady(_: CT.ReadyRequest): CT.ReadyResponse {
    // set error code and finished timestamp immediately if image isn't exist
    if (this.image.image == null) {
      console.error("can not start container without the image");

      this.exit_code = -1;
      this.started_at = getTimestamp();
      this.finished_at = getTimestamp();

      return {
        image: new ArrayBuffer(0),
        args: [],
        envs: {},
      };
    }

    // set started timestamp and pass image to run for web worker
    this.started_at = getTimestamp();
    console.log(this.image.image);
    return {
      image: this.image.image.slice(0),
      args: this.args,
      envs: this.envs,
    };
  }

  _onFinished(request: CT.FinishedRequest): CT.FinishedResponse {
    console.assert(this.finished_at == null);

    this.exit_code = request.code;
    this.finished_at = getTimestamp();

    return {};
  }
}

class SandboxImpl {
  name: string
  uid: string
  namespace: string
  state: PodSandboxState
  containers: Map<string, ContainerImpl>
  created_at: string

  constructor(name: string, uid: string, namespace: string) {
    this.name = name;
    this.uid = uid;
    this.namespace = namespace;
    this.state = PodSandboxState.SandboxReady
    this.containers = new Map<string, ContainerImpl>();
    this.created_at = getTimestamp();
  }

  stop() {
    for (const [_, container] of this.containers) {
      container.stop();
    }
    this.state = PodSandboxState.SandboxNotReady;
    // this.containers.clear();
  }

  createContainer(name: string, image: ImageImpl, args: string[], envs: Record<string, string>): string {
    console.assert(this.state == PodSandboxState.SandboxReady);

    let id: string = (Math.floor(Math.random() * containerIDMax)).toString(16);
    while (containers.has(id)) {
      id = (Math.floor(Math.random() * containerIDMax)).toString(16);
    }

    let container = new ContainerImpl(id, this.uid, name, image, args, envs);
    containers.set(id, container);
    this.containers.set(id, container);

    return id;
  }

  removeContainer(id: string) {
    containers.delete(id);
  }
}

export function initCRI(rootMpx: CL.MultiPlexer): void {
  initHandler(rootMpx);
}

function initHandler(rootMpx: CL.MultiPlexer) {
  let mpx = new CL.MultiPlexer();
  rootMpx.setHandler(crosslinkPath, mpx);

  mpx.setObjHandlerFunc("runPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = runPodSandbox(data as RunPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("stopPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = stopPodSandbox(data as StopPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("removePodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = removePodSandbox(data as RemovePodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("podSandboxStatus", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = podSandboxStatus(data as PodSandboxStatusRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("listPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = listPodSandbox(data as ListPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("createContainer", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = createContainer(data as CreateContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("startContainer", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = startContainer(data as StartContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("stopContainer", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = stopContainer(data as StopContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("removeContainer", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = removeContainer(data as RemoveContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("listContainers", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = listContainers(data as ListContainersRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("containerStatus", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = containerStatus(data as ContainerStatusRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("listImages", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = listImages(data as ListImagesRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("pullImage", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    pullImage(data as PullImageRequest).then((res) => {
      writer.replySuccess(res);
    }).catch((reason) => {
      writer.replyError(reason);
    });
  });

  mpx.setObjHandlerFunc("removeImage", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = removeImage(data as RemoveImageRequest);
    writer.replySuccess(res);
  });
}

// interfaces for CRI

interface RunPodSandboxRequest {
  config: PodSandboxConfig
}

interface PodSandboxConfig {
  metadata: PodSandboxMetadata
}

interface PodSandboxMetadata {
  name: string
  uid: string
  namespace: string
}

interface RunPodSandboxResponse {
  pod_sandbox_id: string
}


interface StopPodSandboxRequest {
  pod_sandbox_id: string
}

interface StopPodSandboxResponse {
  // empty
}

interface RemovePodSandboxRequest {
  pod_sandbox_id: string
}

interface RemovePodSandboxResponse {
  // empty
}

interface PodSandboxStatusRequest {
  pod_sandbox_id: string
}

interface PodSandboxStatusResponse {
  status: PodSandboxStatus
  containers_statuses: ContainerStatus[]
  timestamp: string
}

interface ListPodSandboxRequest {
  filter: PodSandboxFilter
}

interface ListPodSandboxResponse {
  items: PodSandbox[]
}

interface PodSandboxStatus {
  id: string
  metadata: PodSandboxMetadata
  state: PodSandboxState
  created_at: string
}

enum PodSandboxState {
  SandboxReady = 0,
  SandboxNotReady = 1,
}

interface PodSandboxFilter {
  id: string
  state: PodSandboxStateValue
}

interface PodSandboxStateValue {
  state: PodSandboxState
}

interface PodSandbox {
  id: string
  metadata: PodSandboxMetadata
  state: PodSandboxState
  created_at: string
}

interface CreateContainerRequest {
  pod_sandbox_id: string
  config: ContainerConfig
  // sandbox_config: PodSandboxConfig
}

interface ContainerConfig {
  metadata: ContainerMetadata
  image: ImageSpec
  args: string[]
  envs: KeyValue[]
}

interface ContainerMetadata {
  name: string
}

interface CreateContainerResponse {
  container_id: string
}

interface StartContainerRequest {
  container_id: string
}

interface StartContainerResponse {
  // empty
}

interface StopContainerRequest {
  container_id: string
}

interface StopContainerResponse {
  // empty
}

interface RemoveContainerRequest {
  container_id: string
}

interface RemoveContainerResponse {
  // empty
}

interface ListContainersRequest {
  filter: ContainerFilter
}

interface ListContainersResponse {
  containers: Container[]
}

interface ContainerStatusRequest {
  container_id: string
}

interface ContainerStatusResponse {
  status: ContainerStatus
}

interface KeyValue {
  key: string
  value: string
}

interface ContainerStatus {
  id: string
  metadata: ContainerMetadata
  state: ContainerState
  created_at: string
  started_at: string
  finished_at: string
  exit_code: number
  image: ImageSpec
  image_ref: string
}

enum ContainerState {
  ContainerCreated = 0,
  ContainerRunning = 1,
  ContainerExited = 2,
  ContainerUnknown = 3,
}

interface ContainerFilter {
  id: string
  state: ContainerStateValue
  pod_sandbox_id: string
}

interface ContainerStateValue {
  state: ContainerState
}

interface Container {
  id: string
  pod_sandbox_id: string
  metadata: ContainerMetadata
  image: ImageSpec
  image_ref: string
  state: ContainerState
  created_at: string
}

interface ListImagesRequest {
  filter?: ImageFilter
}

interface ImageFilter {
  image: ImageSpec
}

interface ImageSpec {
  image: string
}

interface ListImagesResponse {
  images: Image[]
}

interface Image {
  id: string
  spec: ImageSpec
  // this field meaning the runtime environment of wasm, like 'go:1.19'
  runtime: string
}

interface PullImageRequest {
  image: ImageSpec
}

interface PullImageResponse {
  image_ref: string
}

interface RemoveImageRequest {
  image: ImageSpec
}

interface RemoveImageResponse {
  // nothing
}

// interfaces for communicate with worker 



function runPodSandbox(request: RunPodSandboxRequest): RunPodSandboxResponse {
  // check duplication of name/namespace or uid
  for (const [_, sandbox] of sandboxes) {
    if (sandbox.name === request.config.metadata.name &&
      sandbox.namespace === request.config.metadata.namespace) {
      throw new Error("already exists a sandbox with duplicate name/namespace");
    }
    if (sandbox.uid === request.config.metadata.uid) {
      throw new Error("already exists a sandbox with duplicate uid");
    }
  }

  // assign id of sandbox
  let id: string = (Math.floor(Math.random() * podSandboxIDMax)).toString(16);
  while (sandboxes.has(id)) {
    id = (Math.floor(Math.random() * podSandboxIDMax)).toString(16);
  }

  // allocate new sandbox instance
  let meta: PodSandboxMetadata = request.config.metadata;
  sandboxes.set(id, new SandboxImpl(meta.name, meta.uid, meta.namespace));

  return { pod_sandbox_id: id };
}

function stopPodSandbox(request: StopPodSandboxRequest): StopPodSandboxResponse {
  let sandbox = sandboxes.get(request.pod_sandbox_id);

  if (sandbox != null) {
    sandbox.stop();
  }

  return {};
}

function removePodSandbox(request: RemovePodSandboxRequest): RemovePodSandboxResponse {
  let sandbox = sandboxes.get(request.pod_sandbox_id);

  if (sandbox != null) {
    sandbox.stop();

    // remove all containers of the sandbox
    for (const [_, container] of sandbox.containers) {
      removeContainer({
        container_id: container.id,
      });
    }

    sandboxes.delete(request.pod_sandbox_id);
  }

  return {};
}

function podSandboxStatus(request: PodSandboxStatusRequest): PodSandboxStatusResponse {
  let sandbox = sandboxes.get(request.pod_sandbox_id);

  if (sandbox == null) {
    throw new Error("sandbox not found");
  }

  let containers_statuses: ContainerStatus[] = new Array();
  for (const [_, container] of sandbox.containers) {
    containers_statuses.push({
      id: container.id,
      metadata: {
        name: container.name,
      },
      state: container.getState(),
      created_at: container.created_at,
      started_at: container.started_at ?? "",
      finished_at: container.finished_at ?? "",
      exit_code: container.exit_code ?? 0,
      image: {
        image: container.image.url,
      },
      image_ref: container.image.id,
    });
  }

  return {
    status: {
      id: request.pod_sandbox_id,
      metadata: {
        name: sandbox.name,
        namespace: sandbox.namespace,
        uid: sandbox.uid,
      },
      state: sandbox.state,
      created_at: sandbox.created_at,
    },
    containers_statuses: containers_statuses,
    timestamp: getTimestamp(),
  };
}

function listPodSandbox(request: ListPodSandboxRequest): ListPodSandboxResponse {
  let resSandboxes = new Array<PodSandbox>();
  for (const [id, sandbox] of sandboxes) {
    if (request.filter != null) {
      if (request.filter.id != "" && request.filter.id != id) {
        continue;
      }

      if (request.filter.state != null && request.filter.state.state !=
        sandbox.state) {
        continue;
      }
    }

    resSandboxes.push({
      id: id,
      metadata: {
        name: sandbox.name,
        uid: sandbox.uid,
        namespace: sandbox.namespace,
      },
      state: sandbox.state,
      created_at: sandbox.created_at,
    })
  }

  return { items: resSandboxes };
}

function createContainer(request: CreateContainerRequest): CreateContainerResponse {
  let sandbox = sandboxes.get(request.pod_sandbox_id);
  if (sandbox == null) {
    throw new Error("sandbox not found");
  }

  let image = images.get(request.config.image.image);
  if (image == null) {
    throw new Error("image not found:" + request.config.image.image);
  }

  let args = request.config.args;
  if (args == null) {
    args = new Array<string>();
  }

  let envs: Record<string, string> = {};
  if (request.config.envs != null) {
    for (const env of request.config.envs) {
      envs[env.key] = env.value;
    }
  }

  let id = sandbox.createContainer(request.config.metadata.name, image, args, envs);
  return { container_id: id };
}

function startContainer(request: StartContainerRequest): StartContainerResponse {
  let container = containers.get(request.container_id);
  if (container == null) {
    throw new Error("container not found");
  }
  container.start();
  return {};
}

function stopContainer(request: StopContainerRequest): StopContainerResponse {
  let container = containers.get(request.container_id);
  if (container == null) {
    throw new Error("container not found");
  }
  container.stop();
  return {};
}

function removeContainer(request: RemoveContainerRequest): RemoveContainerResponse {
  let container = containers.get(request.container_id);
  if (container == null) {
    return {};
  }

  container.stop();
  container.cleanup();

  let sandbox = sandboxes.get(container.sandbox_id);
  if (sandbox != null) {
    sandbox.removeContainer(container.id);
  }

  containers.delete(container.id);

  return {};
}

function listContainers(request: ListContainersRequest): ListContainersResponse {
  let resContainers = new Array<Container>();

  for (const [_, container] of containers) {
    if (request.filter != null) {
      let filter = request.filter
      if (filter.id != "" && filter.id !== container.id) {
        continue;
      }

      if (filter.pod_sandbox_id != "" && filter.pod_sandbox_id !== container.sandbox_id) {
        continue;
      }

      if (filter.state != null && filter.state.state !== container.getState()) {
        continue;
      }
    }

    resContainers.push({
      id: container.id,
      pod_sandbox_id: container.sandbox_id,
      metadata: {
        name: container.name,
      },
      image: {
        image: container.image.url,
      },
      image_ref: container.image.id,
      state: container.getState(),
      created_at: container.created_at,
    });
  }

  return { containers: resContainers };
}

function containerStatus(request: ContainerStatusRequest): ContainerStatusResponse {
  let container = containers.get(request.container_id);
  if (container == null) {
    throw new Error("specified container not found");
  }

  return {
    status: {
      id: container.id,
      metadata: {
        name: container.name,
      },
      state: container.getState(),
      created_at: container.created_at,
      started_at: container.started_at || "",
      finished_at: container.finished_at || "",
      exit_code: container.exit_code || 0,
      image: {
        image: container.image.url,
      },
      image_ref: container.image.id,
    }
  };
}

function listImages(request: ListImagesRequest): ListImagesResponse {
  let buf: Array<ImageImpl | undefined> = new Array();
  if (request.filter != null && request.filter.image.image !== "") {
    let image = images.get(request.filter.image.image);
    buf.push(image);

  } else {
    for (const [_, image] of images) {
      if (image.state !== ImageState.Downloaded) {
        continue
      }
      buf.push(image);
    }
  }

  let resImages = new Array<Image>();
  for (const it of buf) {
    if (it == null || it.state != ImageState.Downloaded) {
      continue;
    }

    resImages.push({
      id: it.id,
      spec: {
        image: it.url
      },
      runtime: it.runtime,
    });
  }

  return { images: resImages };
}

function pullImage(request: PullImageRequest): Promise<PullImageResponse> {
  let image = images.get(request.image.image);

  if (image == null) {
    let idSet: Set<string> = new Set();
    for (const [_, image] of images) {
      if (image == null) {
        continue;
      }
      idSet.add(image.id);
    }

    let id: string = (Math.floor(Math.random() * imageRefIDMax)).toString(16);
    while (idSet.has(id)) {
      id = (Math.floor(Math.random() * imageRefIDMax)).toString(16);
    }

    image = new ImageImpl(id, request.image.image);
    images.set(image.url, image);
  }

  console.assert(image.state !== ImageState.Created);

  if (image.state === ImageState.Created ||
    image.state === ImageState.Error) {
    image.pull();
  }

  if (image.state === ImageState.Downloaded) {
    return new Promise<PullImageResponse>((resolve, reject) => {
      if (image == null) {
        reject("logic error on pullImage");
        return;
      }
      resolve({ image_ref: image.id });
    });
  }

  return new Promise<PullImageResponse>((resolve, reject) => {
    let intervalID = setInterval(() => {
      if (image == null) {
        clearInterval(intervalID);
        reject("logic error on pullImage");
        return;
      }

      switch (image.state) {
        case ImageState.Created:
          console.error("logic error");
          break;

        case ImageState.Downloaded:
          resolve({ image_ref: image.id });
          break;

        case ImageState.Downloading:
          // continue to download
          return;

        case ImageState.Error:
          reject("download error on pullImage");
          break;
      }

      clearInterval(intervalID);
    }, 1000);
  });
}

function removeImage(request: RemoveImageRequest): RemoveImageResponse {
  let image = images.get(request.image.image);
  if (image == null) {
    return {};
  }

  if (image.state === ImageState.Downloaded || image.state === ImageState.Error) {
    images.delete(request.image.image)
  }

  return {};
}