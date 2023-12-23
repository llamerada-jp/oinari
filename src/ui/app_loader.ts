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
import * as UTIL from "./util";

interface AppEntry {
  name: string
  description: string
  url: string
}

const applicationsURL = "/applications.json";

const spinnerElementId = "loadAppLibrarySpinner";
const listElementId = "loadAppLibraryList";

let applications: AppEntry[] | undefined;
let downloading: boolean = false;
let command: CMD.Commands;

export function init(cmd: CMD.Commands): void {
  command = cmd;

  let btnLoadApp = document.getElementById("loadAppButton");
  btnLoadApp?.addEventListener("click", show);
}

async function show(): Promise<void> {
  const spinner = document.getElementById(spinnerElementId);
  const list = document.getElementById(listElementId);

  if (applications == undefined) {
    spinner?.classList.remove("d-none");
    list?.classList.add("d-none");

    if (downloading) {
      return;
    }
    await getAppList();
  }

  setupList();
  setupCustom();

  spinner?.classList.add("d-none");
  list?.classList.remove("d-none");
}

async function getAppList(): Promise<void> {
  console.assert(downloading === false);

  downloading = true;
  try {
    let absURL = new URL(applicationsURL, window.location.href);
    let res = await fetch(absURL);
    if (!res.ok) {
      throw new Error('Network response was not OK');
    }
    applications = (await res.json()) as AppEntry[];
    for (let app of applications) {
      app.url = new URL(app.url, absURL).toString();
    }

  } finally {
    downloading = false;
  }
}

function setupList(): void {
  const list = document.getElementById(listElementId);
  if (list == undefined || applications == undefined) {
    return;
  }

  // make empty
  while (list.firstChild) {
    list.removeChild(list.firstChild);
  }

  // add applications
  let temp = document.querySelector("#loadAppLibraryListItem") as HTMLTemplateElement;
  for (const entry of applications) {
    let item = <HTMLElement>temp.content.cloneNode(true);
    (<HTMLElement>item.querySelector(".appName")).innerText = entry.name;
    (<HTMLElement>item.querySelector(".appDescription")).innerText = entry.description;
    (<HTMLElement>item.querySelector(".list-group-item")).addEventListener("click", () => {
      UTIL.closeModal("loadAppClose");
      loadApplication(entry.url);
    });
    list.append(item);
  }
}

function setupCustom(): void {
  const button = document.getElementById("loadAppCustomButton");
  button?.addEventListener("click", () => {
    const urlEl = <HTMLInputElement>document.getElementById("loadAppCustomURL");
    const url = urlEl!.value;
    if (url == "") {
      return;
    }

    UTIL.closeModal("loadAppClose");
    loadApplication(url);
  });
}

function loadApplication(url: string) {
  try {
    command.runApplication(url);
  } catch (e) {
    // TODO: show error using Toasts for the user.
    // https://getbootstrap.com/docs/5.0/components/toasts/
    console.error(e);
  }
}
