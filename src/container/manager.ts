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
import * as Types from "./types";

importScripts("wasm_exec.js");

const CL_PATH: string = Types.CrosslinkPath + "/";

let crosslink: CL.Crosslink;
let rootMpx: CL.MultiPlexer;

export function initManager(cl: CL.Crosslink, rm: CL.MultiPlexer): void {
  crosslink = cl;
  rootMpx = rm;
  initHandler(rootMpx);
}

export function ready(): void {
  crosslink.call(CL_PATH + "ready", {}).then((obj) => {
    let res = obj as Types.ReadyResponse;
    setup(res);
    start(res);
  });
}

function initHandler(rootMpx: CL.MultiPlexer): void {
  let mpx = new CL.MultiPlexer();
  rootMpx.setHandler(Types.CrosslinkPath, mpx);

  mpx.setHandlerFunc("term", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
    // let _ = data as Types.TermRequest;
    // go wasm module can not process term signal, so ignore this
    writer.replySuccess({});
  });
}

function setup(config: Types.ReadyResponse): void {
  for (const r of config.runtime) {
    if (r === "core:dev1") {
      setupCrosslink();
    }
  }
}

function setupCrosslink(): void {
  let goIfHandler = new CL.GoInterfaceHandler();
  goIfHandler.bindCrosslink(crosslink);
  (globalThis as any).crosslink = goIfHandler;
  rootMpx.setDefaultHandler(goIfHandler);
}

function start(config: Types.ReadyResponse): void {
  if (config.image.byteLength === 0) {
    console.log("skip to run program because of empty image");
    return;
  }

  const go = new Go();
  let finFlg = false;

  go.argv = [config.name].concat(config.args);
  go.env = config.envs;
  go.exit = (code: number) => {
    finFlg = true;
    finished(code);
  };

  WebAssembly.instantiate(config.image, go.importObject).then((result) => {
    return go.run(result.instance);

  }).then(() => {
    if (go.exited && !finFlg) {
      finished(0);
    }
  });
}

function finished(code: number): void {
  crosslink.call(CL_PATH + "finished", {
    code: code,
  } as Types.FinishedRequest);
}