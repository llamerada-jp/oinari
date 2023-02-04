import * as CL from "../crosslink"

importScripts("wasm_exec.js");

let rootService: CL.MultiPlexer = new CL.MultiPlexer();
let link: CL.Crosslink = new CL.Crosslink(new class implements CL.WorkerInterface {
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
}(), rootService);

function initService() {
  rootService.setRawHandlerFunc("ping", (_1: string, _2: Map<string, string>, writer: CL.ResponseWriter) => {
    writer.replySuccess("pod");
  });
}

function main() {
  initService();
}

main();