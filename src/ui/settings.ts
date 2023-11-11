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

const allowApplicationsID = "settingsAllowApplications";

let command: CM.Commands;

export function init(cmd: CM.Commands): void {
  command = cmd;

  initAllowApplications();
}

export function setInitConfigs(): void {
  let allowApplications = document.getElementById(allowApplicationsID) as HTMLSelectElement;
  command.setConfiguration(CM.CONFIG_KEY_ALLOW_APPLICATIONS, allowApplications.options[allowApplications.selectedIndex].value);
  command.setConfiguration(CM.CONFIG_KEY_SAMPLE_PREFIX, document.location.origin);
}

function initAllowApplications(): void {
  let allowApplications = document.getElementById(allowApplicationsID) as HTMLSelectElement;
  // set default value
  allowApplications.selectedIndex = 0;

  allowApplications.addEventListener("change", () => {
    let opt = allowApplications.options[allowApplications.selectedIndex];
    let value = opt.value;
    command.setConfiguration(CM.CONFIG_KEY_ALLOW_APPLICATIONS, value);
  });
}