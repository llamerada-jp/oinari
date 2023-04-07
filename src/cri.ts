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

const crosslinkCriPath: string = "cri";
const crosslinkNodePath: string = "node";
const containerTag: string = "container";
const sandboxTag: string = "sandbox";
const podSandboxIdMax: number = Math.floor(Math.pow(2, 30));
const containerIdMax: number = Math.floor(Math.pow(2, 30));
const imageRefIdMax: number = Math.floor(Math.pow(2, 30));
const containerStopTimeout: number = 10 * 1000;

const runtimeRequired: string[] = ["go:1.19"];

let nodeCL: CL.Crosslink;

// key means PodSandboxId
let sandboxes: Map<string, SandboxImpl> = new Map();
// key means ContainerId
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
  image: ArrayBuffer | undefined

  constructor(id: string, url: string) {
    this.state = ImageState.Created;
    this.id = id;
    this.url = url;
    this.pull();
  }

  pull() {
    console.assert(this.state !== ImageState.Downloading, "duplicate download");

    this.state = ImageState.Downloading;

    fetch(this.url).then((response: Response) => {
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
  // PodSandboxId
  sandboxId: string
  name: string
  image: ImageImpl
  worker: Worker | undefined
  link: CL.Crosslink | undefined
  runtime: string[]
  args: string[]
  envs: Record<string, string>
  createdAt: string
  startedAt: string | undefined
  finishedAt: string | undefined
  exitCode: number | undefined

  constructor(id: string, sandboxId: string, name: string, image: ImageImpl, runtime: string[], args: string[], envs: Record<string, string>) {
    this.id = id;
    this.sandboxId = sandboxId;
    this.name = name;
    this.runtime = runtime;
    this.image = image;
    this.args = args;
    this.envs = envs;
    this.createdAt = getTimestamp();
  }

  getState(): ContainerState {
    if (this.finishedAt != null) {
      return ContainerState.ContainerExited;
    }
    if (this.startedAt != null) {
      return ContainerState.ContainerRunning;
    }
    return ContainerState.ContainerCreated;
  }

  start() {
    console.assert(this.worker == null);

    this.startedAt = getTimestamp();
    if (!this._check()) {
      this.exitCode = -1;
      this.finishedAt = getTimestamp();
      return;
    }
    let rootMpx = new CL.MultiPlexer();
    this.worker = new Worker("container.js");
    this.link = new CL.Crosslink(new CL.WorkerImpl(this.worker), rootMpx);

    this._initHandler(rootMpx);
    rootMpx.setHandler(crosslinkNodePath, new NodePipe(this.sandboxId, this.id));
  }

  stop() {
    if (this.worker == null || this.link == null) {
      return;
    }

    this.link.call(CT.CrosslinkPath + "/term", {});
    setTimeout(() => {
      if (this.finishedAt == null) {
        this.finishedAt = getTimestamp();
        // 137 meaning 128 + sig kill(9)
        this.exitCode = 137;
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

  _check(): boolean {
    // image isn't exist
    if (this.image.image == null) {
      console.error("can not start container without the image");
      return false;
    }

    // should specify required runtime
    let hasRequiredRuntime = false;
    for (const r of this.runtime) {
      if (runtimeRequired.indexOf(r) != -1) {
        hasRequiredRuntime = true;
        break;
      }
    }
    if (!hasRequiredRuntime) {
      console.error("minimum required runtime is not specified:" + JSON.stringify(this.runtime));
      return false;
    }

    return true;
  }

  _initHandler(rootMpx: CL.MultiPlexer) {
    let mpx = new CL.MultiPlexer();
    rootMpx.setHandler(CT.CrosslinkPath, mpx);

    mpx.setHandlerFunc("ready", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      let res = this._onReady(data as CT.ReadyRequest);
      writer.replySuccess(res);
    });

    mpx.setHandlerFunc("finished", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      let res = this._onFinished(data as CT.FinishedRequest);
      writer.replySuccess(res);
    });
  }

  _onReady(_: CT.ReadyRequest): CT.ReadyResponse {
    // set error code and finished timestamp immediately if image isn't exist
    if (this.image.image == null) {
      throw new Error("the image should be pulled");
    }

    // pass image to run for web worker
    return {
      name: this.image.url,
      image: this.image.image.slice(0),
      runtime: this.runtime,
      args: this.args,
      envs: this.envs,
    };
  }

  _onFinished(request: CT.FinishedRequest): CT.FinishedResponse {
    console.assert(this.finishedAt == null);

    this.exitCode = request.code;
    this.finishedAt = getTimestamp();

    setTimeout(() => {
      this.cleanup();
    }, containerStopTimeout);
    return {};
  }
}

class SandboxImpl {
  // PodSandboxId
  id: string
  name: string
  // pod.metadata.uid
  uid: string
  namespace: string
  state: PodSandboxState
  containers: Map<string, ContainerImpl>
  createdAt: string

  constructor(id: string, name: string, uid: string, namespace: string) {
    this.id = id;
    this.name = name;
    this.uid = uid;
    this.namespace = namespace;
    this.state = PodSandboxState.SandboxReady
    this.containers = new Map<string, ContainerImpl>();
    this.createdAt = getTimestamp();
  }

  stop() {
    for (const [_, container] of this.containers) {
      container.stop();
    }
    this.state = PodSandboxState.SandboxNotReady;
    // this.containers.clear();
  }

  createContainer(name: string, image: ImageImpl, runtime: string[], args: string[], envs: Record<string, string>): string {
    console.assert(this.state == PodSandboxState.SandboxReady);

    let id: string = (Math.floor(Math.random() * containerIdMax)).toString(16);
    while (this.containers.has(id)) {
      id = (Math.floor(Math.random() * containerIdMax)).toString(16);
    }

    // raise error when there is a container having duplicate name
    for (const [_, container] of this.containers) {
      if (container.name === name) {
        throw new Error("the container having a duplicate name exists");
      }
    }

    let container = new ContainerImpl(id, this.id, name, image, runtime, args, envs);
    containers.set(id, container);
    this.containers.set(id, container);

    return id;
  }

  removeContainer(id: string) {
    this.containers.delete(id);
  }
}

class ApplicationPipe implements CL.Handler {
  serve(data: any, tags: Map<string, string>, writer: CL.ResponseWriter): void {
    let path = tags.get(CL.TAG_PATH);
    let containerId = tags.get(containerTag);
    console.assert(path != null && containerId != null);

    let container = containers.get(containerId!);
    if (container == null) {
      writer.replyError("container isn't exist:" + containerId);
      return;
    }

    container.link?.call(path!, data, tags).then((response) => {
      writer.replySuccess(response);
    }).catch((err) => {
      writer.replyError(err);
    });
  }
}

class NodePipe implements CL.Handler {
  sandboxId: string
  containerId: string

  constructor(sandboxId: string, containerId: string) {
    this.sandboxId = sandboxId;
    this.containerId = containerId;
  }

  serve(data: any, tags: Map<string, string>, writer: CL.ResponseWriter): void {
    let path = tags.get(CL.TAG_PATH);
    console.assert(path != null);

    tags.set(sandboxTag, this.sandboxId);
    tags.set(containerTag, this.containerId);

    nodeCL.call(path!, data, tags).then((response) => {
      writer.replySuccess(response);
    }).catch((err) => {
      writer.replyError(err);
    });
  }
}

export function initCRI(nCL: CL.Crosslink, rootMpx: CL.MultiPlexer): void {
  nodeCL = nCL;
  initHandler(rootMpx);
  rootMpx.setHandler("application", new ApplicationPipe());
}

function initHandler(rootMpx: CL.MultiPlexer) {
  let mpx = new CL.MultiPlexer();

  rootMpx.setHandler(crosslinkCriPath, mpx);

  mpx.setHandlerFunc("runPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = runPodSandbox(data as RunPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("stopPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = stopPodSandbox(data as StopPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("removePodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = removePodSandbox(data as RemovePodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("podSandboxStatus", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = podSandboxStatus(data as PodSandboxStatusRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("listPodSandbox", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = listPodSandbox(data as ListPodSandboxRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("createContainer", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = createContainer(data as CreateContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("startContainer", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = startContainer(data as StartContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("stopContainer", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = stopContainer(data as StopContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("removeContainer", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = removeContainer(data as RemoveContainerRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("listContainers", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = listContainers(data as ListContainersRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("containerStatus", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = containerStatus(data as ContainerStatusRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("listImages", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    let res = listImages(data as ListImagesRequest);
    writer.replySuccess(res);
  });

  mpx.setHandlerFunc("pullImage", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    pullImage(data as PullImageRequest).then((res) => {
      writer.replySuccess(res);
    }).catch((reason) => {
      writer.replyError(reason);
    });
  });

  mpx.setHandlerFunc("removeImage", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
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
  podSandboxId: string
}


interface StopPodSandboxRequest {
  podSandboxId: string
}

interface StopPodSandboxResponse {
  // empty
}

interface RemovePodSandboxRequest {
  podSandboxId: string
}

interface RemovePodSandboxResponse {
  // empty
}

interface PodSandboxStatusRequest {
  podSandboxId: string
}

interface PodSandboxStatusResponse {
  status: PodSandboxStatus
  containersStatuses: ContainerStatus[]
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
  createdAt: string
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
  createdAt: string
}

interface CreateContainerRequest {
  podSandboxId: string
  config: ContainerConfig
  // sandboxConfig: PodSandboxConfig
}

interface ContainerConfig {
  metadata: ContainerMetadata
  image: ImageSpec
  runtime: string[]
  args: string[]
  envs: KeyValue[]
}

interface ContainerMetadata {
  name: string
}

interface CreateContainerResponse {
  containerId: string
}

interface StartContainerRequest {
  containerId: string
}

interface StartContainerResponse {
  // empty
}

interface StopContainerRequest {
  containerId: string
}

interface StopContainerResponse {
  // empty
}

interface RemoveContainerRequest {
  containerId: string
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
  containerId: string
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
  createdAt: string
  startedAt: string
  finishedAt: string
  exitCode: number
  image: ImageSpec
  imageRef: string
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
  podSandboxId: string
}

interface ContainerStateValue {
  state: ContainerState
}

interface Container {
  id: string
  podSandboxId: string
  metadata: ContainerMetadata
  image: ImageSpec
  imageRef: string
  state: ContainerState
  createdAt: string
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
}

interface PullImageRequest {
  image: ImageSpec
}

interface PullImageResponse {
  imageRef: string
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
  let id: string = (Math.floor(Math.random() * podSandboxIdMax)).toString(16);
  while (sandboxes.has(id)) {
    id = (Math.floor(Math.random() * podSandboxIdMax)).toString(16);
  }

  // allocate new sandbox instance
  let meta: PodSandboxMetadata = request.config.metadata;
  sandboxes.set(id, new SandboxImpl(id, meta.name, meta.uid, meta.namespace));

  return { podSandboxId: id };
}

function stopPodSandbox(request: StopPodSandboxRequest): StopPodSandboxResponse {
  let sandbox = sandboxes.get(request.podSandboxId);

  if (sandbox != null) {
    sandbox.stop();
  }

  return {};
}

function removePodSandbox(request: RemovePodSandboxRequest): RemovePodSandboxResponse {
  let sandbox = sandboxes.get(request.podSandboxId);

  if (sandbox != null) {
    sandbox.stop();

    // remove all containers of the sandbox
    for (const [_, container] of sandbox.containers) {
      removeContainer({
        containerId: container.id,
      });
    }

    sandboxes.delete(request.podSandboxId);
  }

  return {};
}

function podSandboxStatus(request: PodSandboxStatusRequest): PodSandboxStatusResponse {
  let sandbox = sandboxes.get(request.podSandboxId);

  if (sandbox == null) {
    throw new Error("sandbox not found");
  }

  let containersStatuses: ContainerStatus[] = new Array();
  for (const [_, container] of sandbox.containers) {
    containersStatuses.push({
      id: container.id,
      metadata: {
        name: container.name,
      },
      state: container.getState(),
      createdAt: container.createdAt,
      startedAt: container.startedAt ?? "",
      finishedAt: container.finishedAt ?? "",
      exitCode: container.exitCode ?? 0,
      image: {
        image: container.image.url,
      },
      imageRef: container.image.id,
    });
  }

  return {
    status: {
      id: request.podSandboxId,
      metadata: {
        name: sandbox.name,
        namespace: sandbox.namespace,
        uid: sandbox.uid,
      },
      state: sandbox.state,
      createdAt: sandbox.createdAt,
    },
    containersStatuses: containersStatuses,
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
      createdAt: sandbox.createdAt,
    })
  }

  return { items: resSandboxes };
}

function createContainer(request: CreateContainerRequest): CreateContainerResponse {
  let sandbox = sandboxes.get(request.podSandboxId);
  if (sandbox == null) {
    throw new Error("sandbox not found");
  }

  let image = images.get(request.config.image.image);
  if (image == null) {
    throw new Error("image not found:" + request.config.image.image);
  }

  let runtime = request.config.runtime;
  if (runtime == null) {
    runtime = new Array<string>();
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

  let id = sandbox.createContainer(request.config.metadata.name, image, runtime, args, envs);
  return { containerId: id };
}

function startContainer(request: StartContainerRequest): StartContainerResponse {
  let container = containers.get(request.containerId);
  if (container == null) {
    throw new Error("container not found");
  }
  container.start();
  return {};
}

function stopContainer(request: StopContainerRequest): StopContainerResponse {
  let container = containers.get(request.containerId);
  if (container == null) {
    throw new Error("container not found");
  }
  container.stop();
  return {};
}

function removeContainer(request: RemoveContainerRequest): RemoveContainerResponse {
  let container = containers.get(request.containerId);
  if (container == null) {
    return {};
  }

  container.stop();
  container.cleanup();

  let sandbox = sandboxes.get(container.sandboxId);
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

      if (filter.podSandboxId != "" && filter.podSandboxId !== container.sandboxId) {
        continue;
      }

      if (filter.state != null && filter.state.state !== container.getState()) {
        continue;
      }
    }

    resContainers.push({
      id: container.id,
      podSandboxId: container.sandboxId,
      metadata: {
        name: container.name,
      },
      image: {
        image: container.image.url,
      },
      imageRef: container.image.id,
      state: container.getState(),
      createdAt: container.createdAt,
    });
  }

  return { containers: resContainers };
}

function containerStatus(request: ContainerStatusRequest): ContainerStatusResponse {
  let container = containers.get(request.containerId);
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
      createdAt: container.createdAt,
      startedAt: container.startedAt || "",
      finishedAt: container.finishedAt || "",
      exitCode: container.exitCode || 0,
      image: {
        image: container.image.url,
      },
      imageRef: container.image.id,
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

    let id: string = (Math.floor(Math.random() * imageRefIdMax)).toString(16);
    while (idSet.has(id)) {
      id = (Math.floor(Math.random() * imageRefIdMax)).toString(16);
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
      resolve({ imageRef: image.id });
    });
  }

  return new Promise<PullImageResponse>((resolve, reject) => {
    let intervalId = setInterval(() => {
      if (image == null) {
        clearInterval(intervalId);
        reject("logic error on pullImage");
        return;
      }

      switch (image.state) {
        case ImageState.Created:
          console.error("logic error");
          break;

        case ImageState.Downloaded:
          resolve({ imageRef: image.id });
          break;

        case ImageState.Downloading:
          // continue to download
          return;

        case ImageState.Error:
          reject("download error on pullImage");
          break;
      }

      clearInterval(intervalId);
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