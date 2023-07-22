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

import * as CM from "../command";

let command: CM.Commands;
let processing: boolean = false;
let account: string;
let node: string;
let spinners = ["procListByAccountSpinner1", "procListByAccountSpinner2"];

export function init(cmd: CM.Commands): void {
  command = cmd;

  let btnProcList = document.getElementById("procListButton");
  btnProcList?.addEventListener("click", reload);
  let btnProcRefresh = document.getElementById("procListRefresh");
  btnProcRefresh?.addEventListener("click", reload);
}

export function setNodeInfo(a: string, n: string) {
  account = a;
  node = n;
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
    list?.classList.add("d-none");
    while (list?.firstChild) {
      list.removeChild(list.firstChild);
    }
  }

  // get process list
  let procList = await command.listProcess()

  // add list items
  for (let proc of procList) {
    let content = new Map<string, string>();
    content.set(".appName", proc.name);
    content.set(".appPhase", proc.phase);
    content.set(".appOwnerAccount", proc.owner);
    content.set(".appRunningNode", proc.runningNode);

    if (proc.owner === account && listByAccount != null) {
      let item = addListItem(listByAccount, temp, content);
      item.querySelector(".appMenuTerminate")?.addEventListener("click", () => {
        command.terminateProcess(proc.uuid);
      });
    }
    if (proc.runningNode === node && listByNode != null) {
      let item = addListItem(listByNode, temp, content);
      item.querySelector(".appMenuTerminate")?.addEventListener("click", () => {
        command.terminateProcess(proc.uuid);
      });
    }
  }

  // hide spinner and show list
  for (let id of spinners) {
    const spinner = document.getElementById(id);
    spinner?.classList.add("d-none");
  }
  for (let list of [listByAccount, listByNode]) {
    list?.classList.remove("d-none");
  }

  processing = false;
}

function addListItem(list: HTMLElement, temp: HTMLTemplateElement, contents: Map<string, string>): HTMLElement {
  let item = temp.content.cloneNode(true) as HTMLElement;
  for (const [key, value] of contents) {
    (item.querySelector(key) as HTMLElement).innerText = value;
  }
  list.append(item);
  return item;
}