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
import * as POS from "../position";

const deviceInfoElID = "deviceInfo";
const accountElID = "deviceInfoAccount";
const deviceNameElID = "deviceInfoDeviceName";
const positionElID = "deviceInfoPosition";

export function init(localSettings: LS.LocalSettings, position: POS.Position): void {
  setText(accountElID, localSettings.account);
  setText(deviceNameElID, localSettings.deviceName);
  updatePosition(position.coordinate);
  position.addListener((coordinate: POS.Coordinate) => {
    updatePosition(coordinate);
  });
}

export function show(): void {
  let el = document.getElementById(deviceInfoElID) as HTMLElement;
  el.classList.remove("d-none");
}

function updatePosition(coordinate: POS.Coordinate): void {
  setText(positionElID, `${coordinate.latitude}, ${coordinate.longitude}`);
}

function setText(elID: string, text:string):void{
  let el = document.getElementById(elID) as HTMLElement;
  el.innerText = text;
}