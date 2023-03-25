import * as CL from "./crosslink";

class WorkerMock implements CL.WorkerInterface {
  pair!: WorkerMock;
  listener!: (datum: any) => void;

  addEventListener(listener: (datum: any) => void): void {
    this.listener = listener;
  }

  post(datum: object): void {
    this.pair.listener(datum);
  }

  setPair(pair: WorkerMock) {
    this.pair = pair;
  }
}

class HandlerMock implements CL.Handler {
  data: any;
  tags: Map<string, string>;

  constructor() {
    this.tags = new Map();
  }

  serve(data: any, tags: Map<string, string>, writer: CL.ResponseWriter): void {
    this.data = data;
    this.tags = tags;
    writer.replySuccess("response");
  }
}

test("handler", () => {
  const mock1 = new WorkerMock();
  const mock2 = new WorkerMock();
  mock1.setPair(mock2);
  mock2.setPair(mock1);

  const crosslink1 = new CL.Crosslink(mock1, new HandlerMock());
  let handler = new HandlerMock();
  const crosslink2 = new CL.Crosslink(mock2, handler);

  crosslink1.call(
    "test_path",
    { key: "value" },
    new Map<string, string>([["tag", "tag content"]])

  ).then((response: any) => {
    expect(response).toBe("response");
    expect(handler.data.key).toBe("value");
    expect(handler.tags.size).toBe(2);
    expect(handler.tags.get(CL.TAG_PATH)).toBe("test_path");
    expect(handler.tags.get("tag")).toBe("tag content");
  });
});

describe("response", () => {
  const mock1 = new WorkerMock();
  const mock2 = new WorkerMock();
  mock1.setPair(mock2);
  mock2.setPair(mock1);

  const mpx = new CL.MultiPlexer();
  const crosslink1 = new CL.Crosslink(mock1, mpx);
  const crosslink2 = new CL.Crosslink(mock2, new HandlerMock());

  mpx.setHandlerFunc("success", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    let o = data as {
      request: string,
    };
    expect(o.request).toBe("request success");
    expect(tags.get("tag")).toBe("tag success");
    writer.replySuccess({
      res: "reply success"
    });
  });

  mpx.setHandlerFunc("failure", (_1: any, _2: Map<string, string>, writer: CL.ResponseWriter) => {
    writer.replyError("reply failure");
  });

  mpx.setHandlerFunc("exception", (_1: any, _2: Map<string, string>, writer: CL.ResponseWriter) => {
    throw "exception reply";
  });

  test("func with success reply", () => {
    return crosslink2.call("success", {
      request: "request success",
    }, new Map<string, string>([["tag", "tag success"]])).then((response: any) => {
      let r = response as { res: string };
      expect(r.res).toBe("reply success");
    });
  });

  test("func with error message", () => {
    return crosslink2.call("failure", null, new Map<string, string>()).catch((message) => {
      expect(message).toBe("reply failure");
    });
  });

  test("func with exception", () => {
    return crosslink2.call("exception", null, new Map<string, string>()).catch((message) => {
      expect(message).toBe("exception: exception reply");
    });
  });
});

test("multiplexer", () => {
  const mock1 = new WorkerMock();
  const mock2 = new WorkerMock();
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

  mpxRoot.setHandlerFunc("func1", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    expect(data).toBe("request func1");
    expect(tags.get(CL.TAG_PATH)).toBe("func1");
    writer.replyError("reply func1");
  });

  mpxRoot.setHandler("branch", mpxBranch);

  mpxBranch.setHandlerFunc("func2", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    let param = data as {
      message: string
    }
    expect(param.message).toBe("request func2");
    expect(tags.get(CL.TAG_PATH)).toBe("branch/func2");
    writer.replySuccess({
      message: "reply func2"
    });
  });

  crosslink1.call("notexist", "request default").then((response) => {
    expect(response).toBe("reply default");
  }).catch(() => {
    throw "unreachable";
  });

  crosslink1.call("func1", "request func1").then(() => {
    throw "unreachable";
  }).catch((response) => {
    expect(response).toBe("reply func1");
  })

  crosslink1.call("branch/func2", {
    message: "request func2"
  }).then((response) => {
    let r = response as {
      message: string
    };
    expect(r.message).toBe("reply func2");
  }).catch(() => {
    throw "unreachable";
  });

  crosslink1.call("branch/func2/dummy", "").then(() => {
    throw "unreachable";
  }).catch((message) => {
    expect(message).toBe("handler not found. path:branch/func2/dummy");
  })

  crosslink1.call("branch/dummy", "").then(() => {
    throw "unreachable";
  }).catch((message) => {
    expect(message).toBe("handler not found. path:branch/dummy");
  });
});