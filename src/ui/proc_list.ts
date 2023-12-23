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

import * as CMD from "../command";
import * as LS from "../local_settings";
import * as MI from "./migrate";
import * as UTIL from "./util";

let command: CMD.Commands;
let processing: boolean = false;
let account: string;
let nodeID: string;
let spinners = ["procListByAccountSpinner1", "procListByAccountSpinner2"];

export function init(cmd: CMD.Commands, localSettings: LS.LocalSettings, nID: string): void {
  command = cmd;
  account = localSettings.account;
  nodeID = nID;

  let btnProcList = document.getElementById("procListButton");
  btnProcList?.addEventListener("click", reload);
  let btnProcRefresh = document.getElementById("procListRefresh");
  btnProcRefresh?.addEventListener("click", reload);
}

async function reload(): Promise<void> {
  if (processing) {
    return;
  }
  processing = true;

  // show spinner
  for (let id of spinners) {
    const spinner = document.getElementById(id);
    spinner?.classList.remove("d-none");
  }

  let listByAccount = document.getElementById("procListByAccountList");
  let listByNode = document.getElementById("procListInNodeList");
  let temp = document.getElementById("procListItem") as HTMLTemplateElement;

  // make list empty
  for (let list of [listByAccount, listByNode]) {
    UTIL.makeListEmptyHide(list);
  }

  // make nodeID - name map
  let nodeList = await command.listNode();
  let nodeMap = new Map<string, string>();
  for (let node of nodeList) {
    nodeMap.set(node.id, node.name);
  }

  // get process list
  let procList = await command.listProcess();

  // add list items
  for (let proc of procList) {
    let node = proc.runningNodeID;
    if (nodeMap.has(node)) {
      node = nodeMap.get(node)!;
    }
    let content = new Map<string, string | UTIL.clickEventCB>();
    content.set(".appName", proc.name);
    content.set(".appState", proc.state);
    content.set(".appOwnerAccount", proc.owner);
    content.set(".appRunningNode", node);
    content.set(".appMenuTerminate", () => {
      command.terminateProcess(proc.uuid);
    });
    content.set(".appMenuMigrate", () => {
      UTIL.closeModal("procListClose");
      MI.showMigrateModal(proc.uuid, proc.runningNodeID);
    });

    if (proc.owner === account && listByAccount != null) {
      UTIL.addListItem(listByAccount, temp, content);
    }
    if (proc.runningNodeID === nodeID && listByNode != null) {
      UTIL.addListItem(listByNode, temp, content);
    }
  }

  // hide spinner and show list
  for (let id of spinners) {
    const spinner = document.getElementById(id);
    spinner?.classList.add("d-none");
  }
  for (let list of [listByAccount, listByNode]) {
    UTIL.makeListShow(list);
  }

  processing = false;
}
