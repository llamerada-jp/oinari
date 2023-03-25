import * as CL from "./crosslink";

export class PodInterface {
  private localId: string;
  private worker: Worker | null;
  private service: CL.MultiPlexer;
  private link: CL.Crosslink | null;

  constructor(localId: string) {
    this.localId = localId;
    this.worker = null;
    this.service = new CL.MultiPlexer();
    this.link = null;
  }

  run() {
    let worker: Worker = new Worker("pod.js");
    this.worker = worker;
    this.link = new CL.Crosslink(new CL.WorkerImpl(worker), this.service);
  }

  stop() {
    if (this.worker === null) {
      return;
    }

    this.worker.terminate();
    this.worker = null;
    this.link = null;
  }

  getLink(): CL.Crosslink | undefined {
    if (this.link === null) {
      return undefined;
    }

    return this.link;
  }

  getLocalId(): string {
    return this.localId;
  }
}

export class PodManager {
  private workers: Map<string, PodInterface>;

  constructor() {
    this.workers = new Map<string, PodInterface>();
  }

  create(): PodInterface {
    let localPodId: string;
    do {
      localPodId = Math.floor(Math.random() * 2147483647).toString(16);
    } while (this.workers.has(localPodId));

    let pi: PodInterface = new PodInterface(localPodId);
    pi.run();

    return pi;
  }

  remove(podLocalId: string) {
    let pi = this.workers.get(podLocalId);
    if (pi === undefined) {
      return;
    }

    pi.stop();
    this.workers.delete(podLocalId);
  }

  get(podLocalId: string): PodInterface | undefined {
    return this.workers.get(podLocalId);
  }

  list(): Map<string, PodInterface> {
    return this.workers;
  }
}

export function initPodManagerService(manager: PodManager, rootService: CL.MultiPlexer) {
  let mpx = new CL.MultiPlexer();
  rootService.setHandler("podManager", mpx);

  mpx.setHandlerFunc("create", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    let pi = manager.create();
    writer.replySuccess({
      localPodId: pi.getLocalId(),
    });
  });

  mpx.setHandlerFunc("delete", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    let param = (data as {
      localPodId: string
    });
    manager.remove(param.localPodId);
  });
}

export function initPodService(manager: PodManager, rootService: CL.MultiPlexer) {
  rootService.setHandlerFunc("pod", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    let localPodId = tags.get("localPodId");
    if (localPodId === undefined) {
      writer.replyError("should specify `localPodId`");
      return;
    }

    let pi = manager.get(localPodId);
    if (pi === undefined) {
      writer.replyError("pod is not exist: " + localPodId);
      return;
    }

    let path = tags.get(CL.TAG_LEAF);
    if (path === undefined || path === "") {
      writer.replyError("logic error");
      return;
    }
    tags.set(CL.TAG_PATH, path);
    tags.delete(CL.TAG_LEAF);

    pi.getLink()?.call(tags.get(CL.TAG_PATH) || "", data, tags).then((response) => {
      writer.replySuccess(response);
    }).catch((reason) => {
      writer.replyError(reason);
    })
  });
}