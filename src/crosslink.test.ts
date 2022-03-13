import * as CL from "./crosslink";

class workerMock implements CL.WorkerInterface {
    pair!: workerMock;
    listener!: (datum: any) => void;

    addEventListener(listener: (datum: any) => void): void {
        this.listener = listener;
    }

    post(datum: object): void {
        this.pair.listener(datum);
    }

    setPair(pair: workerMock) {
        this.pair = pair;
    }
}

class handlerMock implements CL.Handler {
    serve(data: string, tags: Map<string, string>, writer: CL.ResponseWriter): void {
        switch (tags.get("func")) {
            case "success":
                writer.replySuccess("success reply");
                break;

            case "failure":
                writer.replyError("failure reply");
                break;

            case "exception":
                throw "exception reply";

            default:
                throw "unreachable";

        }
    }
}

describe('crosslink', () => {
    const mock1 = new workerMock();
    const mock2 = new workerMock();
    mock1.setPair(mock2);
    mock2.setPair(mock1);

    const crosslink1 = new CL.Crosslink(mock1, new handlerMock());
    const crosslink2 = new CL.Crosslink(mock2, new handlerMock());

    test("func with success reply", () => {
        return crosslink1.call("success", new Map<string, string>([["func", "success"]])).then((reply) => {
            expect(reply).toBe("success reply");
        });
    });

    test("func with error message", () => {
        return crosslink2.call("failure", new Map<string, string>([["func", "failure"]])).catch((message) => {
            expect(message).toBe("failure reply");
        });
    });

    test("func with exception", () => {
        return crosslink2.call("exception", new Map<string, string>([["func", "exception"]])).catch((message) => {
            expect(message).toBe("exception: exception reply");
        });
    });
});

function makeTags(path: string): Map<string, string> {
    let r = new Map<string, string>();
    r.set(CL.TAG_PATH, path);
    return r;

}

describe("multiplexer", () => {
    const mock1 = new workerMock();
    const mock2 = new workerMock();
    mock1.setPair(mock2);
    mock2.setPair(mock1);

    const mpxRoot = new CL.MultiPlexer();
    const mpxBranch = new CL.MultiPlexer();

    const crosslink1 = new CL.Crosslink(mock1, new class implements CL.Handler {
        serve(_1: string, _2: Map<string, string>, _3: CL.ResponseWriter): void {
            throw "unreachable";
        }
    });

    const crosslink2 = new CL.Crosslink(mock2, mpxRoot);
    mpxRoot.setDefaultHandler(new class implements CL.Handler {
        serve(data: string, tags: Map<string, string>, writer: CL.ResponseWriter): void {
            expect(data).toBe("request default");
            expect(tags.get(CL.TAG_PATH)).toBe("notexist");
            writer.replySuccess("reply default");
        }
    });

    mpxRoot.setHandlerFunc("func1", (data: string, tags: Map<string, string>, writer: CL.ResponseWriter) => {
        expect(data).toBe("request func1");
        expect(tags.get(CL.TAG_PATH)).toBe("func1");
        writer.replyError("reply func1");
    });

    mpxRoot.setHandler("branch", mpxBranch);

    mpxBranch.setHandlerFunc("func2", (data: string, tags: Map<string, string>, writer: CL.ResponseWriter) => {
        expect(data).toBe("request func2");
        expect(tags.get(CL.TAG_PATH)).toBe("branch/func2");
        writer.replySuccess("reply func2");
    });

    crosslink1.call("request default", makeTags("notexist")).then((reply) => {
        expect(reply).toBe("reply default");
    }).catch(() => {
        throw "unreachable";
    });

    crosslink1.call("request func1", makeTags("func1")).then(() => {
        throw "unreachable";
    }).catch((reply) => {
        expect(reply).toBe("reply func1");
    })

    crosslink1.call("request func2", makeTags("branch/func2")).then((reply) => {
        expect(reply).toBe("reply func2");
    }).catch(() => {
        throw "unreachable";
    });

    crosslink1.call("", makeTags("branch/func2/dummy")).then(() => {
        throw "unreachable";
    }).catch((message) => {
        expect(message).toBe("handler not found. path:branch/func2/dummy");
    })

    crosslink1.call("", makeTags("branch/dummy")).then(()=>{
        throw "unreachable";
    }).catch((message) => {
        expect(message).toBe("handler not found. path:branch/dummy");
    });
});