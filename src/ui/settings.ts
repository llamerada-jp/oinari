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

import * as LS from "../local_settings";
import * as UTIL from "./util";
import * as DEF from "../definitions"

const settingsLinkID = "settingsLink";
const settingsSubmitButtonID = "settingsSubmit";

const deviceNameElID = "settingsDeviceName";
const viewTypeElID = "settingsViewType";
const localStoreInputID = "settingsLocalStore";
const syncGNSSInputID = "settingsSyncGNSS";
const spawnPositionInputID = "settingsSpawnPosition";
const allowApplicationsInputID = "settingsAllowApplications";

let localSettings: LS.LocalSettings;
let onSubmit: () => void;

export function init(ls: LS.LocalSettings, submit: () => void): void {
  localSettings = ls;
  onSubmit = submit;

  initElements();
}

export function initElements(): void {
  let settingsButton = document.getElementById(settingsLinkID);
  settingsButton?.addEventListener("click", loadSettings);

  let settingsCloseButton = document.getElementById(settingsSubmitButtonID);
  settingsCloseButton?.addEventListener("click", () => {
    storeSettings();
    onSubmit();
  });

  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  let spawnPositionInput = document.getElementById(spawnPositionInputID) as HTMLInputElement;

  if (!!navigator.geolocation) {
    syncGNSSInput?.addEventListener("click", () => {
      if (syncGNSSInput!.checked) {
        spawnPositionInput!.disabled = true;
      } else {
        spawnPositionInput!.disabled = false;
      }
    });

  } else {
    syncGNSSInput!.disabled = true;
    syncGNSSInput!.checked = false;
    spawnPositionInput!.disabled = false;
  }
}

function loadSettings(): void {
  let deviceNameEl = document.getElementById(deviceNameElID) as HTMLInputElement;
  deviceNameEl!.innerText = localSettings.deviceName;

  let viewTypeEl = document.getElementById(viewTypeElID) as HTMLSelectElement;
  switch (localSettings.viewType) {
    case DEF.VIEW_TYPE_LANDSCAPE:
      viewTypeEl!.innerText = "landscape";
      break;

    case DEF.VIEW_TYPE_XR:
      viewTypeEl!.innerText = "XR";
      break;
  }

  let localStoreInput = document.getElementById(localStoreInputID) as HTMLInputElement;
  localStoreInput!.checked = localSettings.enableLocalStore;

  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  syncGNSSInput!.checked = localSettings.enableGNSS;
  syncGNSSInput!.dispatchEvent(new Event("click"));

  if (localSettings.position) {
    let spawnPositionSelect = document.getElementById(spawnPositionInputID) as HTMLSelectElement;
    UTIL.setSelectValue(spawnPositionSelect, localSettings.position);
  }

  let allowApplicationsInput = document.getElementById(allowApplicationsInputID) as HTMLSelectElement;
  UTIL.setSelectValue(allowApplicationsInput!, localSettings.allowApplications);
}

function storeSettings(): void {
  let localStoreInput = document.getElementById(localStoreInputID) as HTMLInputElement;
  localSettings.enableLocalStore = localStoreInput!.checked;

  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  localSettings.enableGNSS = syncGNSSInput!.checked;

  if (!localSettings.enableGNSS) {
    let spawnPositionSelect = document.getElementById(spawnPositionInputID) as HTMLSelectElement;
    let positionStr = spawnPositionSelect.options[spawnPositionSelect.selectedIndex].value;
    localSettings.position = positionStr;
  }

  let allowApplicationsSelect = document.getElementById(allowApplicationsInputID) as HTMLSelectElement;
  localSettings.allowApplications = allowApplicationsSelect.options[allowApplicationsSelect.selectedIndex].value;
}