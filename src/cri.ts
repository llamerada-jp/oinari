import * as CL from "./crosslink";

const crosslinkPath: string = "cri";
const podSandboxIDMax: number = Math.floor(Math.pow(2, 30));
const containerIDMax: number = Math.floor(Math.pow(2, 30));
const imageRefIDMax: number = Math.floor(Math.pow(2, 30));

// key means PodSandboxID
let sandboxes: Map<string, Sandbox> = new Map();
// key means ContainerID
let containers: Map<string, Container> = new Map();
// key means image url (not image ref id)
let imageByURL: Map<string, ImageInstance> = new Map();
let images: Array<WeakRef<ImageInstance>> = new Array();

function getTimestamp(): string {
  return new Date().toISOString();
}

enum ImageState {
  Created = 0,
  Downloading = 1,
  Downloaded = 2,
  Error = 3,
}

class ImageInstance {
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
    this.reload();
  }

  reload() {
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

      return fetch(param.image);

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

class Container {
  id: string
  // PodSandboxID
  sandbox_id: string
  name: string
  image: ImageInstance
  worker: Worker | undefined
  link: CL.Crosslink | undefined
  service: CL.MultiPlexer
  created_at: string
  started_at: string | undefined
  finished_at: string | undefined
  exit_code: number | undefined

  constructor(id: string, sandbox_id: string, name: string, image: ImageInstance) {
    this.id = id;
    this.sandbox_id = sandbox_id;
    this.name = name;
    this.image = image;
    this.service = new CL.MultiPlexer();
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

    this.worker = new Worker("container.js");
    this.link = new CL.Crosslink(new CL.WorkerImpl(this.worker), this.service);

    console.error("implement to start program!");
  }

  stop() {
    if (this.worker == null) {
      return;
    }

    this.worker.terminate();
    this.worker = undefined;
    this.link = undefined;
  }
}

class Sandbox {
  name: string
  uid: string
  namespace: string
  containers: Map<string, Container>
  created_at: string

  constructor(name: string, uid: string, namespace: string) {
    this.name = name;
    this.uid = uid;
    this.namespace = namespace;
    this.containers = new Map<string, Container>();
    this.created_at = getTimestamp();
  }

  stop() {
    for (const [_, container] of this.containers) {
      container.stop();
    }
    this.containers.clear();
  }

  createContainer(name: string, image: ImageInstance): string {
    let id: string = (Math.floor(Math.random() * containerIDMax)).toString(16);
    while (containers.has(id)) {
      id = (Math.floor(Math.random() * containerIDMax)).toString(16);
    }

    let container = new Container(id, this.uid, name, image);
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

  mpx.setObjHandlerFunc("listImages", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = listImages(data as ListImagesRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("pullImage", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = pullImage(data as PullImageRequest);
    writer.replySuccess(res);
  });

  mpx.setObjHandlerFunc("removeImage", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
    let res = removeImage(data as RemoveImageRequest);
    writer.replySuccess(res);
  });
}

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

interface CreateContainerRequest {
  pod_sandbox_id: string
  config: ContainerConfig
  // sandbox_config: PodSandboxConfig
}

interface ContainerConfig {
  metadata: ContainerMetadata
  image: ImageSpec
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

function runPodSandbox(request: RunPodSandboxRequest): RunPodSandboxResponse {
  let id: string = (Math.floor(Math.random() * podSandboxIDMax)).toString(16);
  while (sandboxes.has(id)) {
    id = (Math.floor(Math.random() * podSandboxIDMax)).toString(16);
  }

  let meta: PodSandboxMetadata = request.config.metadata;
  sandboxes.set(id, new Sandbox(meta.name, meta.uid, meta.namespace));

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
      id: sandbox.uid,
      metadata: {
        name: sandbox.name,
        namespace: sandbox.namespace,
        uid: sandbox.uid,
      },
      state: PodSandboxState.SandboxReady,
      created_at: sandbox.created_at,
    },
    containers_statuses: containers_statuses,
    timestamp: getTimestamp(),
  };
}

function createContainer(request: CreateContainerRequest): CreateContainerResponse {
  let sandbox = sandboxes.get(request.pod_sandbox_id);
  if (sandbox == null) {
    throw new Error("sandbox not found");
  }

  let image = imageByURL.get(request.config.image.image);
  if (image == null) {
    throw new Error("image not found:" + request.config.image.image);
  }

  let id = sandbox.createContainer(request.config.metadata.name, image);
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

  let sandbox = sandboxes.get(container.sandbox_id);
  if (sandbox != null) {
    sandbox.removeContainer(container.id);
  }

  return {};
}

function listImages(request: ListImagesRequest): ListImagesResponse {
  let buf: Array<ImageInstance | undefined> = new Array();
  if (request.filter != null) {
    let image = imageByURL.get(request.filter.image.image);
    buf.push(image);

  } else {
    for (const it of images) {
      buf.push(it.deref());
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

function pullImage(request: PullImageRequest): PullImageResponse {
  let idSet: Set<string> = new Set();
  for (const it of images) {
    let instance = it.deref();
    if (instance == null) {
      continue;
    }
    idSet.add(instance.id);
  }

  let id: string = (Math.floor(Math.random() * imageRefIDMax)).toString(16);
  while (idSet.has(id)) {
    id = (Math.floor(Math.random() * imageRefIDMax)).toString(16);
  }

  let image = new ImageInstance(id, request.image.image);
  imageByURL.set(image.url, image);
  images.push(new WeakRef<ImageInstance>(image));

  return { image_ref: id };
}

function removeImage(request: RemoveImageRequest): RemoveImageResponse {
  let url = request.image.image;
  imageByURL.delete(url);

  images = images.filter((it): boolean => {
    let instance = it.deref();
    if (instance == null) {
      return false;
    }
    return instance.url !== url;
  });

  return {};
}