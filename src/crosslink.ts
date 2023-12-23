type Waiting = {
  resolve: (value: any | PromiseLike<any>) => void;
  reject: (reason?: any) => void;
};

export const TAG_PATH: string = "path";
export const TAG_LEAF: string = "leaf";

export interface WorkerInterface {
  addEventListener(listener: (datum: any) => void): void;
  post(datum: object): void;
}

export class WorkerImpl implements WorkerInterface {
  worker: Worker;
  listener: (datum: object) => void;

  constructor(worker: Worker) {
    this.worker = worker;
    this.listener = (_: object) => { }; // init by dummy

    this.worker.addEventListener("message", (event) => {
      this.listener(event.data);
    });
  }

  addEventListener(listener: (datum: object) => void): void {
    this.listener = listener;
  }

  post(datum: object): void {
    this.worker.postMessage(datum);
  }
}

export class CoWorkerImpl implements WorkerInterface {
  listener: (datum: object) => void;

  constructor() {
    this.listener = (_: object) => { }; // init by temporary dummy

    globalThis.addEventListener("message", (event) => {
      this.listener(event.data);
    })
  }

  addEventListener(listener: (datum: object) => void): void {
    this.listener = listener;
  }

  post(datum: object): void {
    globalThis.postMessage(datum);
  }
}

export class ResponseWriter {
  private id: number;
  private worker: WorkerInterface;
  private replied: boolean;

  constructor(id: number, worker: WorkerInterface) {
    this.id = id;
    this.worker = worker;
    this.replied = false;
  }

  replySuccess(response: any): void {
    if (this.replied) {
      console.error("response replied yet.");
      return;
    }

    this.worker.post({
      type: "response",
      id: this.id,
      response: response,
    });

    this.replied = true;
  }

  replyError(message: string): void {
    if (this.replied) {
      console.error("response replied yet.");
      return;
    }

    this.worker.post({
      type: "error",
      id: this.id,
      message: message,
    });

    this.replied = true;
  }

  isReplied(): boolean {
    return this.replied;
  }
}

export interface Handler {
  serve(data: any, tags: Map<string, string>, writer: ResponseWriter): void;
}

export class Crosslink {
  private worker: WorkerInterface;
  private handler: Handler;
  private waitings: Map<number, Waiting>;

  constructor(worker: WorkerInterface, handler: Handler) {
    this.worker = worker;
    this.worker.addEventListener((datum: {
      type: string;
      id: number;
      tags: Map<string, string>, // for call
      data: any, // for call
      response: any, // for response
      message: string, // for error
    }) => {
      switch (datum.type) {
        case "call":
          this.receiveCall(datum.data, datum.tags, datum.id);
          break;

        case "response":
          this.receiveResponse(datum.response, datum.id);
          break;

        case "error":
          this.receiveError(datum.message, datum.id);
          break;

        default:
          console.error("receive unsupported message", datum);
          break;
      }
    });

    this.handler = handler;

    this.waitings = new Map<number, Waiting>();
  }

  call(path: string, param: any, tags?: Map<string, string>): Promise<any> {
    return new Promise((resolve, reject) => {
      let copyTag: Map<string, string> = new Map<string, string>();
      // copy tags to avoid changing original object
      if (tags != null) {
        for (const [key, value] of tags) {
          copyTag.set(key, value);
        }
      }
      copyTag.set(TAG_PATH, path);

      let id = this.assignId();
      this.waitings.set(id, {
        resolve: resolve,
        reject: reject,
      });
      this.worker.post({
        type: "call",
        data: param,
        tags: copyTag,
        id: id,
      });
    });
  }

  private assignId(): number {
    let id: number;
    do {
      id = Math.floor(Math.random() * 2147483647);
    } while (this.waitings.has(id));
    return id;
  }

  private receiveCall(data: any, tags: Map<string, string>, id: number): void {
    let writer = new ResponseWriter(id, this.worker);

    try {
      this.handler.serve(data, tags, writer);

    } catch (e) {
      if (!writer.isReplied()) {
        this.worker.post({
          type: "error",
          id: id,
          message: `exception: ${e}`,
        });
      }
      console.error("exception", e);
    }
  }

  private receiveResponse(response: any, id: number): void {
    let waiting = this.waitings.get(id);

    if (waiting === undefined) {
      console.error("logic error");
      return;
    }

    this.waitings.delete(id);
    waiting.resolve(response);
  }

  private receiveError(message: string, id: number): void {
    let waiting = this.waitings.get(id);

    if (waiting === undefined) {
      console.error("logic error");
      return;
    }

    this.waitings.delete(id);
    waiting.reject(message);
  }
}

const MULTI_PLEXOR_SPLITTER = /^\/?([^\/]*)\/?(.*)$/;

export class MultiPlexer implements Handler {
  private defaultHandler: Handler;
  private handlers: Map<string, Handler>;

  constructor() {
    this.defaultHandler = new class implements Handler {
      serve(_: any, tags: Map<string, string>, writer: ResponseWriter): void {
        writer.replyError("handler not found. path:" + tags.get(TAG_PATH));
      }
    };
    this.handlers = new Map<string, Handler>();
  }

  serve(data: any, tags: Map<string, string>, writer: ResponseWriter): void {
    let path = tags.get(TAG_PATH);
    if (path === undefined) {
      writer.replyError("`path` tag should be set.");
      return;
    }

    let leaf = tags.get(TAG_LEAF)
    if (leaf === undefined) {
      leaf = path
    }

    let r = MULTI_PLEXOR_SPLITTER.exec(leaf)

    let dir: string | undefined;
    let newLeaf: string | undefined;
    if (r !== null) {
      dir = r[1];
      newLeaf = r[2];
    }
    if (newLeaf === undefined) {
      newLeaf = "";
    }

    let handler = this.defaultHandler;
    let newTags = new Map<string, string>(tags);
    newTags.set(TAG_LEAF, newLeaf);

    if (dir != undefined) {
      let namedHandler = this.handlers.get(dir);
      if (namedHandler !== undefined) {
        handler = namedHandler;
      }
    }

    handler.serve(data, newTags, writer)
  }

  setHandler(pattern: string, handler: Handler): void {
    this.handlers.set(pattern, handler);
  }

  setHandlerFunc(pattern: string, f: (data: any, tags: Map<string, string>, writer: ResponseWriter) => void) {
    this.handlers.set(pattern, new class implements Handler {
      serve(data: any, tags: Map<string, string>, writer: ResponseWriter): void {
        if (!tags.has(TAG_LEAF) || tags.get(TAG_LEAF) !== "") {
          writer.replyError("handler not found. path:" + tags.get(TAG_PATH));
          return;
        }

        f(data, tags, writer);
      }
    });
  }

  setDefaultHandler(handler: Handler): void {
    this.defaultHandler = handler;
  }
}

export class GoInterfaceHandler implements Handler {
  private cl: Crosslink | undefined;
  private rwMap: Map<number, ResponseWriter>;

  constructor() {
    this.rwMap = new Map<number, ResponseWriter>();
  }

  bindCrosslink(cl: Crosslink): void {
    this.cl = cl;
  }

  serve(data: any, tags: Map<string, string>, writer: ResponseWriter): void {
    const ID_MAX: number = Math.pow(2, 31)
    let id: number = 0;
    const rwMap = this.rwMap;

    do {
      id = Math.floor(Math.random() * ID_MAX);
    } while (rwMap.has(id));
    rwMap.set(id, writer);

    let jsTags: Record<string, string> = {};
    for (const key of tags.keys()) {
      let value = tags.get(key);
      if (key !== TAG_LEAF && value !== undefined) {
        jsTags[key] = value;
      }
    }

    this.serveToGo(id, JSON.stringify(data), JSON.stringify(jsTags));
  }

  serveToGo(id: number, data: string, tags: string): void {
    console.assert(false, "this method will be override by go");
  }

  replyFromGo(id: number, response: string, message: string): void {
    let rw = this.rwMap.get(id);

    if (rw === undefined) {
      console.assert(false, "the assigned id must be exist");
      return;
    }
    this.rwMap.delete(id);

    if (message !== "") {
      rw.replyError(message);
    } else {
      rw.replySuccess(JSON.parse(response));
    }
  }

  callFromGo(id: number, path: string, data: string, tags: string) {
    if (this.cl === undefined) {
      console.error("Crosslink should be bind on setup.");
      return;
    }

    let mapTags = new Map<string, string>()

    if (tags != null && tags != "") {
      let jsTags = JSON.parse(tags) as Record<string, string>;
      for (const key of Object.keys(jsTags)) {
        mapTags.set(key, jsTags[key]);
      }
    }

    let obj: any = null
    if (data != null && data != "") {
      obj = JSON.parse(data);
    }

    this.cl.call(path, obj, mapTags).then((response) => {
      this.replyToGo(id, JSON.stringify(response), "");
    }).catch((message: string) => {
      this.replyToGo(id, "", message);
    });
  }

  replyToGo(_1: number, _2: string, _3: string): void {
    console.assert(false, "this method will be override by go");
  }
}