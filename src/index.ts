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
import * as LS from "./local_settings";
import * as POS from "./position";
import * as UI_AL from "./ui/app_loader";
import * as UI_IS from "./ui/init_settings";
import * as UI_MAP from "./ui/map";
import * as UI_MI from "./ui/migrate";
import * as UI_PL from "./ui/proc_list";
import * as UI_SE from "./ui/settings";
import * as UI_SI from "./ui/system_info";

declare function ColonioModule(): Promise<any>;

let rootMpx: CL.MultiPlexer;
let frontendMpx: CL.MultiPlexer;

let crosslink: CL.Crosslink;
let command: CM.Commands;
let localSettings: LS.LocalSettings;
let position: POS.Position;

export function main(account: string): void {
  // start controller
  initController().then(() => {
    command = new CM.Commands(crosslink);

    localSettings = new LS.LocalSettings(command, account);
    position = new POS.Position(command);

    // init ui after window loaded
    if (document.readyState !== "loading") {
      initializeSettings(false);
    } else {
      document.addEventListener("DOMContentLoaded", () => {
        initializeSettings(false);
      }, false);
    }
  });
}

// This function will be called from google apis callback.
export function readyMap(): void {
  UI_MAP.readyMap();
}

async function initController(): Promise<void> {
  // start controller worker
  const controller = new Worker("controller.js");

  // setup crosslink
  rootMpx = new CL.MultiPlexer();
  crosslink = new CL.Crosslink(new CL.WorkerImpl(controller), rootMpx);

  // setup CRI
  CRI.initCRI(crosslink, rootMpx);

  frontendMpx = new CL.MultiPlexer();
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

function initializeSettings(retry: boolean): void {
  if (!retry) {
    UI_IS.init(localSettings, () => {
      UI_IS.hide();
      connect();
    });
  }

  // skip initializeSettings if already initialized
  if (localSettings.enableLocalStore && !retry) {
    UI_IS.hide();
    connect();
    return;
  }

  UI_IS.show();
}

async function connect(): Promise<void> {
  // try to close when close window or tab
  window.addEventListener('beforeunload', function () {
    command.disconnect();
  });

  try {
    let connectInfo = await command.connect(
      "https://localhost:8080/seed",
      localSettings.account,
      "",
      localSettings.deviceName,
      "PC");

    localSettings.apply();

    // set position
    position.enableGNSS = localSettings.enableGNSS;
    if (!localSettings.enableGNSS) {
      position.setCoordinateByStr(localSettings.position);
    }

    // publish node info
    await command.setPublicity(10.0);

    start(connectInfo);

  } catch (error) {
    console.error(error);
    initializeSettings(true);
  }
}

function start(connectInfo: CM.ConnectInfo): void {
  switch (localSettings.viewType) {
    case "landscape":
      startLandscape(connectInfo);
      break;
    default:
      throw new Error("Unknown view type: " + localSettings.viewType);
  }
}

function startLandscape(connectInfo: CM.ConnectInfo): void {
  UI_AL.init(command);
  UI_MI.init(command);
  UI_PL.init(command, localSettings, connectInfo.nodeID);
  UI_SE.init(localSettings, () => {
    localSettings.apply();

    position.enableGNSS = localSettings.enableGNSS;
    if (!localSettings.enableGNSS) {
      position.setCoordinateByStr(localSettings.position);
    }
  });
  UI_SI.init(localSettings, position);

  UI_MAP.init(frontendMpx, position, () => {
    // show
    UI_SI.show();
    UI_MAP.show();
    let menuEl = document.getElementById("menu") as HTMLDivElement;
    menuEl.classList.remove("d-none");
  });
}