import * as CL from "./crosslink";
import * as misc from "./crosslink.test.misc";

export class Tester {
  private crosslink: CL.Crosslink;
  private crosslinkGo: CL.GoInterfaceHandler;
  private called: number = 0;

  // setup modules
  constructor() {
    const workerMock1 = new misc.workerMock();
    const workerMock2 = new misc.workerMock();

    workerMock1.setPair(workerMock2);
    workerMock2.setPair(workerMock1);

    const mpx = new CL.MultiPlexer();

    this.crosslink = new CL.Crosslink(workerMock1, mpx);
    let goIfHandler = new CL.GoInterfaceHandler();
    let crosslink = new CL.Crosslink(workerMock2, goIfHandler);
    goIfHandler.bindCrosslink(crosslink);
    this.crosslinkGo = goIfHandler;

    // setup handlers
    mpx.setRawHandlerFunc("jsFunc", (data: string, tags: Map<string, string>, writer: CL.ResponseWriter) => {
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
      const wasm = fetch("./test_crosslink.wasm");
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
    this.crosslink.callRaw("request go", misc.makeTags("goFunc")).then((value: string) => {
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