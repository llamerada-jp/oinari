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

import * as CL from "./crosslink";
import * as CM from "./command";
import * as WB from "./webrtc_bypass_handler";
import * as CRI from "./cri";
import * as UI_AL from "./ui/app_loader";
import * as UI_MAP from "./ui/map";
import * as UI_MI from "./ui/migrate";
import * as UI_PL from "./ui/proc_list";
import * as UI_SI from "./ui/system_info";

declare function ColonioModule(): Promise<any>;

let rootMpx: CL.MultiPlexer;
let crosslink: CL.Crosslink;
let command: CM.Commands;

async function initController(): Promise<void> {
  // start controller worker
  const controller = new Worker("controller.js");

  // setup crosslink
  rootMpx = new CL.MultiPlexer();
  crosslink = new CL.Crosslink(new CL.WorkerImpl(controller), rootMpx);

  // setup CRI
  CRI.initCRI(crosslink, rootMpx);

  let frontendMpx = new CL.MultiPlexer();
  rootMpx.setHandler("frontend", frontendMpx);

  // setup colonio module handler
  let colonio = await ColonioModule();
  let webrtcImpl: WebrtcImplement = new colonio.DefaultWebrtcImplement();
  frontendMpx.setHandler("webrtc", WB.NewWebrtcHandler(crosslink, webrtcImpl));

  // setup frontend module handler
  let promise = new Promise<void>((resolve) => {
    frontendMpx.setHandlerFunc("nodeReady", (_1: any, _2: Map<string, string>, writer: CL.ResponseWriter) => {
      writer.replySuccess("");
      resolve();
    });
  });

  // run wasm program of controller
  crosslink.call("run", {
    file: "oinari.wasm",
  });

  return promise;
}

function initUI() {
  UI_AL.init(command);
  UI_MAP.init();
  UI_MI.init(command);
  UI_PL.init(command);
}

async function main(): Promise<void> {
  // start controller
  await initController();
  command = new CM.Commands(crosslink);

  // init ui after window loaded
  if (document.readyState !== "loading") {
    initUI();
  } else {
    document.addEventListener("DOMContentLoaded", initUI, false);
  }

  // try to close when close window or tab
  window.addEventListener('beforeunload', function () {
    command.disconnect();
  });

  // connect
  // let acTmp = Math.random().toString(32).substring(2);
  let account = "account-1";
  let nodeName = "node-" + Math.random().toString(32).substring(2);
  let info = await command.connect("https://localhost:8080/seed",
    account,
    "",
    nodeName,
    "PC");
  UI_SI.set(info.account, info.node);
  UI_PL.setNodeInfo(info.account, info.node);

  // set a position for sample playing
  await command.setPosition(35.6594945, 139.6999859, 0);
  // publish node info
  await command.setPublicity(10.0);
}

main();