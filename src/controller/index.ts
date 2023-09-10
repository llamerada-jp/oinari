/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

  go.exit = (code: number) => {
    if (code === 0) {
      writer.replySuccess({});
    } else {
      writer.replyError("controller process failed");
    }
  }

  ColonioModule().then((colonio) => {
    // bypass webrtc
    let bypass = new WB.WebrtcBypass(crosslink, rootMpx, "frontend/webrtc");
    colonio.setWebrtcImpl(bypass);

    // setup colonio for go
    (globalThis as any).colonioGo = new ColonioGo(colonio);

    return fetch(req.file);

  }).then((response)=> {
    return WebAssembly.instantiateStreaming(response, go.importObject);

  }).then((wasm) => {
    // start go program
    go.run(wasm.instance);
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