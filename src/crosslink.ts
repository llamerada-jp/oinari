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

export class ResponseWriter {
    private id: number;
    private worker: WorkerInterface;
    private replied: boolean;

    constructor(id: number, worker: WorkerInterface) {
        this.id = id;
        this.worker = worker;
        this.replied = false;
    }

    replySuccess(result: string): void {
        if (this.replied) {
            console.error("response replied yet.");
            return;
        }

        this.worker.post({
            type: "reply",
            id: this.id,
            result: result,
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
    notEdge?: boolean
    serve(data: string, tags: Map<string, string>, writer: ResponseWriter): void;
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
            data: string, // for call
            result: any, // for reply
            message: string, // for error
        }) => {
            switch (datum.type) {
                case "call":
                    this.receiveCall(datum.data, datum.tags, datum.id);
                    break;

                case "reply":
                    this.receiveReply(datum.result, datum.id);
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

    call(data: string, tags: Map<string, string>): Promise<string> {
        return new Promise((resolve, reject) => {
            let id = this.assignId();
            this.waitings.set(id, {
                resolve: resolve,
                reject: reject,
            });
            this.worker.post({
                type: "call",
                data: data,
                tags: tags,
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

    private receiveCall(data: string, tags: Map<string, string>, id: number): void {
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

    private receiveReply(result: object, id: number): void {
        let waiting = this.waitings.get(id);

        if (waiting === undefined) {
            console.error("logic error");
            return;
        }

        this.waitings.delete(id);
        waiting.resolve(result);
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

const MULTI_PLEXER_SPLITER = /^\/?([^\/]*)\/?(.*)$/;

export class MultiPlexer implements Handler {
    notEdge?: boolean = true;

    private defaultHandler: Handler;
    private handlers: Map<string, Handler>;

    constructor() {
        this.defaultHandler = new class implements Handler {
            serve(_: string, tags: Map<string, string>, writer: ResponseWriter): void {
                writer.replyError("handler not found. path:" + tags.get(TAG_PATH));
            }
        };
        this.handlers = new Map<string, Handler>();
    }

    serve(data: string, tags: Map<string, string>, writer: ResponseWriter): void {
        let path = tags.get(TAG_PATH);
        if (path === undefined) {
            writer.replyError("`path` tag should be set.");
            return;
        }

        let leaf = tags.get(TAG_LEAF)
        if (leaf === undefined) {
            leaf = path
        }

        let r = MULTI_PLEXER_SPLITER.exec(leaf)

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
                // handler is edge and leaf string is empty
                if (namedHandler.notEdge === undefined && newLeaf === "") {
                    newTags.delete(TAG_LEAF);
                    handler = namedHandler;
                }

                // handler is not edge
                if (namedHandler.notEdge != undefined) {
                    handler = namedHandler;
                }
            }
        }

        handler.serve(data, newTags, writer)
    }

    setHandler(pattern: string, handler: Handler): void {
        this.handlers.set(pattern, handler);
    }

    setHandlerFunc(pattern: string, f: (data: string, tags: Map<string, string>, writer: ResponseWriter) => void) {
        this.handlers.set(pattern, new class implements Handler {
            serve(data: string, tags: Map<string, string>, writer: ResponseWriter): void {
                f(data, tags, writer);
            }
        })
    }

    setDefaultHandler(handler: Handler): void {
        this.defaultHandler = handler;
    }
}

interface stringMap {
    [key: string]: string;
}

export class CrosslinkGo {
    private cl: Crosslink;
    private rwMap: Map<number, ResponseWriter>;

    constructor(worker: WorkerInterface) {
        let handler = new class implements Handler {
            notEdge?: boolean = true;
            private clGo: CrosslinkGo;

            constructor(clGo: CrosslinkGo) {
                this.clGo = clGo;
            }

            serve(data: string, tags: Map<string, string>, writer: ResponseWriter): void {
                let id: number = 0;
                const rwMap = this.clGo.rwMap;

                do {
                    id = Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
                } while (rwMap.has(id));
                rwMap.set(id, writer);

                let jsTags: stringMap = {};
                for (const key of tags.keys()) {
                    let value = tags.get(key);
                    if (value !== undefined) {
                        jsTags[key] = value;
                    }
                }

                this.clGo.serveToGo(id, data, JSON.stringify(jsTags));
            }
        }(this);

        this.rwMap = new Map<number, ResponseWriter>();
        this.cl = new Crosslink(worker, handler);
    }

    serveToGo(id: number, data: string, tags: string): void {
        console.assert(false, "this method will be override by go");
    }

    serveReplyFromGo(id: number, result: string, message: string): void {
        let rw = this.rwMap.get(id);

        if (rw === undefined) {
            console.assert(false, "the assigned id must be exist");
            return;
        }
        this.rwMap.delete(id);

        if (message !== "") {
            rw.replyError(message);
        } else {
            rw.replySuccess(result);
        }
    }

    callFromGo(id: number, data: string, tags: string) {
        let jsTags = JSON.parse(tags) as stringMap;
        let mapTags = new Map<string, string>()

        for (const key of Object.keys(jsTags)) {
            mapTags.set(key, jsTags[key]);
        }

        this.cl.call(data, mapTags).then((reply) => {
            this.callReplyToGo(id, reply, "");
        }).catch((message: string) => {
            this.callReplyToGo(id, "", message);
        });
    }

    callReplyToGo(id: number, reply: string, message: string): void {
        console.assert(false, "this method will be override by go");
    }
}