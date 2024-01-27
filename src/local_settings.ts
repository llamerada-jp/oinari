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

import * as CMD from "./command";
import * as DEF from "./definitions"

const LOCAL_STORAGE_KEY_PREFIX = "localSettings";

const ALLOW_APPLICATIONS_SAMPLES = "samples";
const ALLOW_APPLICATIONS_MYSELF = "myself";
const ALLOW_APPLICATIONS_SAMPLES_AND_MYSELF = "samplesAndMyself";
const ALLOW_APPLICATIONS_ANY = "any";
const ALLOW_APPLICATIONS = [
  ALLOW_APPLICATIONS_SAMPLES, ALLOW_APPLICATIONS_MYSELF,
  ALLOW_APPLICATIONS_SAMPLES_AND_MYSELF, ALLOW_APPLICATIONS_ANY
];

interface LocalSettingsData {
  account: string;
  allowApplications: string;
  deviceName: string;
  enableGNSS: boolean;
  position: string;
  viewType: string;
}

export class LocalSettings {
  private localStorageKey: string;
  private cmd: CMD.Commands;
  private settings: LocalSettingsData;
  enableLocalStore: boolean;

  constructor(cmd: CMD.Commands, account: string) {
    this.localStorageKey = LOCAL_STORAGE_KEY_PREFIX + ":" + account;
    this.cmd = cmd;
    this.enableLocalStore = false;

    this.settings = {} as LocalSettingsData;

    if (localStorage.getItem(this.localStorageKey)) {
      // load settings from local storage
      this.enableLocalStore = true;
      this.loadFromLocalStorage(account);
    }

    // set default settings
    this.fillDefaultSettings(account);
  }

  apply() {
    // store settings to local storage
    if (this.enableLocalStore) {
      this.saveToLocalStorage();
    } else {
      localStorage.removeItem(this.localStorageKey);
    }

    this.cmd.setConfiguration(CMD.CONFIG_KEY_ALLOW_APPLICATIONS, this.settings.allowApplications);
    this.cmd.setConfiguration(CMD.CONFIG_KEY_SAMPLE_PREFIX, document.location.origin);
  }

  private loadFromLocalStorage(account: string) {
    if (!this.enableLocalStore) {
      throw new Error("local storage is not enabled");
    }

    let raw = localStorage.getItem(this.localStorageKey);
    if (raw == null || raw === "") {
      throw new Error("invalid local storage data");
    }

    let settings = JSON.parse(raw) as LocalSettingsData;
    // remove local storage if account is changed
    if (settings.account !== account) {
      localStorage.removeItem(this.localStorageKey);
      this.settings = {} as LocalSettingsData;
      return;
    }

    this.settings = settings;
  }

  private fillDefaultSettings(account: string) {
    if (this.settings.account === undefined) {
      this.settings.account = account;
    }

    if (this.settings.allowApplications === undefined) {
      this.settings.allowApplications = ALLOW_APPLICATIONS_SAMPLES;
    }

    if (this.settings.deviceName === undefined) {
      this.settings.deviceName = "";
    }

    if (this.settings.enableGNSS === undefined) {
      this.settings.enableGNSS = false;
    }

    if (this.settings.viewType === undefined) {
      this.settings.viewType = DEF.VIEW_TYPE_LANDSCAPE;
    }
  }

  private saveToLocalStorage() {
    if (!this.enableLocalStore) {
      throw new Error("local storage is not enabled");
    }

    let raw = JSON.stringify(this.settings);
    localStorage.setItem(this.localStorageKey, raw);
  }

  get account(): string {
    return this.settings.account;
  }

  get allowApplications(): string {
    return this.settings.allowApplications;
  }

  set allowApplications(filter: string) {
    if (!ALLOW_APPLICATIONS.includes(filter)) {
      throw new Error("invalid allow applications");
    }

    this.settings.allowApplications = filter;
  }

  get deviceName(): string {
    return this.settings.deviceName;
  }

  set deviceName(name: string) {
    this.settings.deviceName = name;
  }

  get enableGNSS(): boolean {
    return this.settings.enableGNSS;
  }

  set enableGNSS(enable: boolean) {
    this.settings.enableGNSS = enable;
  }

  get position(): string {
    return this.settings.position;
  }

  set position(pos: string) {
    this.settings.position = pos;
  }

  get viewType(): string {
    return this.settings.viewType;
  }

  set viewType(type: string) {
    if (!DEF.VIEW_TYPES.includes(type)) {
      throw new Error("invalid view type");
    }

    this.settings.viewType = type;
  }
}