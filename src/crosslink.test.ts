import * as CL from "./crosslink";
import * as misc from "./crosslink.test.misc";

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
  const mock1 = new misc.workerMock();
  const mock2 = new misc.workerMock();
  mock1.setPair(mock2);
  mock2.setPair(mock1);

  const crosslink1 = new CL.Crosslink(mock1, new handlerMock());
  const crosslink2 = new CL.Crosslink(mock2, new handlerMock());

  test("func with success reply", () => {
    return crosslink1.callRaw("success", new Map<string, string>([["func", "success"]])).then((reply) => {
      expect(reply).toBe("success reply");
    });
  });

  test("func with error message", () => {
    return crosslink2.callRaw("failure", new Map<string, string>([["func", "failure"]])).catch((message) => {
      expect(message).toBe("failure reply");
    });
  });

  test("func with exception", () => {
    return crosslink2.callRaw("exception", new Map<string, string>([["func", "exception"]])).catch((message) => {
      expect(message).toBe("exception: exception reply");
    });
  });
});


describe("multiplexer", () => {
  const mock1 = new misc.workerMock();
  const mock2 = new misc.workerMock();
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

  mpxRoot.setRawHandlerFunc("func1", (data: string, tags: Map<string, string>, writer: CL.ResponseWriter) => {
    expect(data).toBe("request func1");
    expect(tags.get(CL.TAG_PATH)).toBe("func1");
    writer.replyError("reply func1");
  });

  mpxRoot.setHandler("branch", mpxBranch);

  mpxBranch.setObjHandlerFunc("func2", (data: any, tags: Map<string, string>, writer: CL.ResponseObjWriter) => {
    let param = data as {
      message: string
    }
    expect(param.message).toBe("request func2");
    expect(tags.get(CL.TAG_PATH)).toBe("branch/func2");
    writer.replySuccess({
      message: "reply func2"
    });
  });

  crosslink1.callRaw("request default", misc.makeTags("notexist")).then((reply) => {
    expect(reply).toBe("reply default");
  }).catch(() => {
    throw "unreachable";
  });

  crosslink1.callRaw("request func1", misc.makeTags("func1")).then(() => {
    throw "unreachable";
  }).catch((reply) => {
    expect(reply).toBe("reply func1");
  })

  crosslink1.callRaw(JSON.stringify({
    message: "request func2"
  }), misc.makeTags("branch/func2")).then((reply) => {
    let r = JSON.parse(reply) as {
      message: string
    };
    expect(r.message).toBe("reply func2");
  }).catch(() => {
    throw "unreachable";
  });

  crosslink1.callRaw("", misc.makeTags("branch/func2/dummy")).then(() => {
    throw "unreachable";
  }).catch((message) => {
    expect(message).toBe("handler not found. path:branch/func2/dummy");
  })

  crosslink1.callRaw("", misc.makeTags("branch/dummy")).then(() => {
    throw "unreachable";
  }).catch((message) => {
    expect(message).toBe("handler not found. path:branch/dummy");
  });
});