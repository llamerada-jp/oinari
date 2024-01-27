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

export interface Coordinate {
  latitude: number;
  longitude: number;
  altitude: number;
};

export class Position {
  private cmd: CMD.Commands;
  private watchID: number;
  private listeners: ((coordinate: Coordinate) => void)[];
  private _enableGNSS: boolean;
  private _coordinate: Coordinate;

  constructor(cmd: CMD.Commands) {
    this.cmd = cmd;
    this.watchID = 0;
    this.listeners = [];
    this._enableGNSS = false;
    this._coordinate = {} as Coordinate;
  }

  addListener(listener: (coordinate: Coordinate) => void): void {
    this.listeners.push(listener);
  }

  applyPosition(): void {
    this.cmd.setPosition({
      x: this._coordinate.longitude,
      y: this._coordinate.latitude,
      z: this._coordinate.altitude,
    });
    for (let listener of this.listeners) {
      listener(this._coordinate);
    }
  }

  get enableGNSS(): boolean {
    return this._enableGNSS;
  }

  set enableGNSS(enable: boolean) {
    if (enable && !navigator.geolocation) {
      throw new Error("This device does not support GNSS.");
    }
    this._enableGNSS = enable;
  }

  watchGNSS() {
    if (!this._enableGNSS) {
      throw new Error("GNSS is not enabled.");
    }
    // watching GNSS yet.
    if (this.watchID !== 0) {
      return;
    }

    this.watchID = navigator.geolocation.watchPosition((position) => {
      let altitude = 0;
      if (position.coords.altitude) {
        altitude = position.coords.altitude;
      }
      this._coordinate = {
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
        altitude: altitude,
      };
      this.applyPosition();

    }, (error) => {
      console.error(error);

    }, {
      enableHighAccuracy: true,
      maximumAge: 0,
      timeout: 10000,
    });
  }

  unwatchGNSS() {
    if (this.watchID === 0) {
      return;
    }

    navigator.geolocation.clearWatch(this.watchID);
    this.watchID = 0;
  }

  get coordinate(): Coordinate {
    return this._coordinate;
  }

  set coordinate(c: Coordinate) {
    if (this._enableGNSS) {
      throw new Error("GNSS is not enabled.");
    }
    this._coordinate = c;
  }

  setCoordinateByStr(coordinate: string): void {
    if (this._enableGNSS) {
      throw new Error("GNSS is not enabled.");
    }

    let tmp = coordinate.split(",");
    if (tmp.length !== 2) {
      throw new Error("invalid coordinate string: " + coordinate);
    }

    this.coordinate = {
      latitude: this.generateRandomNumByRange(tmp[0]),
      longitude: this.generateRandomNumByRange(tmp[1]),
      altitude: 0,
    }
  }

  private generateRandomNumByRange(range: string): number {
    let rangeArr = range.split("-");
    if (rangeArr.length === 1) {
      return parseFloat(rangeArr[0]);
    }

    if (rangeArr.length !== 2) {
      throw new Error("invalid range parameter: " + range);
    }

    let min = parseFloat(rangeArr[0]);
    let max = parseFloat(rangeArr[1]);
    if (max > min) {
      throw new Error("invalid range parameter: " + range);
    }
    return Math.random() * (max - min) + min;
  }
}