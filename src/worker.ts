import * as CL from "./crosslink";

importScripts("wasm_exec.js");

// Setup crosslink.
const workerInterface = new class implements CL.WorkerInterface {
  listener: (datum:object) => void;

  constructor() {
    this.listener = (_:object) => {
      // dummy
    }

    addEventListener("message", (event) => {
      this.listener(event.data);
    })
  }

  addEventListener(listener: (datum: object) => void): void {
    this.listener = listener;
  }

  post(datum: object): void {
    postMessage(datum);
  }
}();
const crosslink = new CL.Crosslink(workerInterface);

// Start go program.
const go = new Go();
const wasm = fetch("./oinari.wasm");

WebAssembly.instantiateStreaming(wasm, go.importObject).then(result => {
  go.run(result.instance);
});