import * as CL from "../crosslink";
import * as WB from "./webrtc_bypass";

importScripts("colonio.js", "colonio_go.js", "wasm_exec.js");

declare function ColonioModule(): Promise<any>;

// setup crosslink
let rootMpx = new CL.MultiPlexer();
let crosslink = new CL.Crosslink(new CL.CoWorkerImpl(), rootMpx);

// load method
interface RunRequest {
  file: string
}

function run(data: any, _: Map<string, string>, writer: CL.ResponseWriter): void {
  let req = data as RunRequest;

  const go = new Go();
  const wasm = fetch(req.file);

  ColonioModule().then((colonio) => {
    // bypass webrtc
    let colonioMpx = new CL.MultiPlexer();
    rootMpx.setHandler("colonio", colonioMpx);
    let bypass = new WB.WebrtcBypass(crosslink, colonioMpx);
    colonio.setWebrtcImpl(bypass);

    // setup colonio for go
    (globalThis as any).colonioGo = new ColonioGo(colonio);

    return WebAssembly.instantiateStreaming(wasm, go.importObject);

  }).then((result) => {
    // start go program
    go.run(result.instance);
    writer.replySuccess({});
  });
}

function main() {
  let goIfHandler = new CL.GoInterfaceHandler();
  goIfHandler.bindCrosslink(crosslink);
  (globalThis as any).crosslink = goIfHandler;
  rootMpx.setDefaultHandler(goIfHandler);
  rootMpx.setHandlerFunc("run", run);
}

main();