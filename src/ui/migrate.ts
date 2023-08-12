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
import * as Util from "./util";

let command: CM.Commands;
let processing: boolean = false;
let uuid: string;

const spinnerID = "migrateListSpinner";
const listID = "migrateList";
const tempID = "migrateListItem";

export function init(cmd: CM.Commands): void {
  command = cmd;

  let btnRefresh = document.getElementById("migrateRefresh");
  btnRefresh?.addEventListener("click", reload);
}

export function showMigrateModal(u: string): void {
  uuid = u;
  reload();
}

async function reload(): Promise<void> {
  if (processing) {
    return;
  }
  processing = true;

  // show spinner
  const spinnerEl = document.getElementById(spinnerID);
  spinnerEl?.classList.remove("d-none");
  // make list empty
  const listEl = document.getElementById(listID);
  if (listEl == null) {
    console.error("element not found", listID);
    return;
  }
  Util.makeListEmptyHide(listEl);

  let temp = document.getElementById(tempID) as HTMLTemplateElement;
  let nodeList = await command.listNode();
  for (let node of nodeList) {
    let content = new Map<string, string | Util.clickEventCB>();
    content.set(".nodeName", node.name);
    content.set(".nodeID", node.id);
    content.set(".nodeType", node.nodeType);
    content.set(".nodeLatitude", node.latitude.toString());
    content.set(".nodeLongitude", node.longitude.toString());
    content.set(".nodeAltitude", node.altitude.toString());
    Util.addListItem(listEl, temp, content);
  }

  // hide spinner and show list
  spinnerEl?.classList.add("d-none");
  Util.makeListShow(listEl);

  processing = false;
}