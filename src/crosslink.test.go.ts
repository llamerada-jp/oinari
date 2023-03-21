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

export class Tester {
  private crosslink: CL.Crosslink;
  private crosslinkGo: CL.GoInterfaceHandler;
  private called: number = 0;

  // setup modules
  constructor() {
    const workerMock1 = new WorkerMock();
    const workerMock2 = new WorkerMock();

    workerMock1.setPair(workerMock2);
    workerMock2.setPair(workerMock1);

    const mpx = new CL.MultiPlexer();

    this.crosslink = new CL.Crosslink(workerMock1, mpx);
    let goIfHandler = new CL.GoInterfaceHandler();
    let crosslink = new CL.Crosslink(workerMock2, goIfHandler);
    goIfHandler.bindCrosslink(crosslink);
    this.crosslinkGo = goIfHandler;

    // setup handlers
    mpx.setHandlerFunc("jsFunc", (data: any, tags: Map<string, string>, writer: CL.ResponseWriter) => {
      this.called++;
      console.assert(data === "request js");
      console.assert(tags.get(CL.TAG_PATH) === "jsFunc");

      if (tags.get("type") === "success") {
        writer.replySuccess("result js success");
      } else {
        writer.replyError("result js failure");
      }
    });
  }

  // start test
  async start(): Promise<boolean> {
    try {
      this.called = 0;
      // set global values for testing
      (globalThis as any).crosslinkGo = this.crosslinkGo;
      (globalThis as any).crosslinkGoTest = this;

      // start go program
      const go = new Go();
      const wasm = fetch("./test/test_crosslink.wasm");
      const instance = await WebAssembly.instantiateStreaming(wasm, go.importObject)
      await go.run(instance.instance);

      console.assert(this.called == 2);
    } finally {
      // cleanup
      delete (globalThis as any).crosslinkGo
      delete (globalThis as any).crosslinkGoTest
    }

    return true;
  }

  // called by go
  runByGo() {
    this.crosslink.call("goFunc", "request go").then((value: string) => {
      console.assert(value === "result go func1");
      this.finToGo(true);
    }).catch(() => {
      console.assert(false, "unreachable in this test case");
    })
  }

  finToGo(result: boolean) {
    console.assert(false, "this method will be override by go")
  }
}