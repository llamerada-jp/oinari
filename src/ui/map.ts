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

import * as CL from "../crosslink";
import * as POS from "../position";
import * as V from "./view";

import * as THREE from "three";
import { ThreeJSOverlayView } from "@googlemaps/three";
import { Keys } from "../keys";

const mainViewElID = "mainView";

let overlayView: OinariOverlayView;
let position: POS.Position;
let hadReady: boolean = false;
let onReady: () => void;

const mapOptions = {
  tilt: 67.5,
  heading: 0,
  zoom: 18,
  center: { lat: 35.6594945, lng: 139.6999859 },
  mapId: Keys.googleMapID,
  // disable interactions due to animation loop and moveCamera
  disableDefaultUI: true,
  gestureHandling: "none",
  keyboardShortcuts: false,
};

export function init(frontendMpx: CL.MultiPlexer, pos: POS.Position, ready: () => void): void {
  initHandler(frontendMpx);
  position = pos;

  if (hadReady) {
    ready();
    start();
  } else {
    onReady = ready;
  }
}

export function readyMap(): void {
  hadReady = true;
  if (onReady) {
    onReady();
    start();
  }
}

export function show(): void {
  let viewEl = document.getElementById(mainViewElID);
  viewEl!.classList.remove("d-none");
}

function start(): void {
  overlayView = new OinariOverlayView(position);
}

function initHandler(frontendMpx: CL.MultiPlexer): void {
  frontendMpx.setHandlerFunc("applyObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as V.ApplyObjectsRequest;
    overlayView.applyObjects(request.objects);
    writer.replySuccess(null);
  });

  frontendMpx.setHandlerFunc("deleteObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as V.DeleteObjectsRequest;
    overlayView.deleteObjects(request.uuids);
    writer.replySuccess(null);
  });
}

class OinariOverlayView extends ThreeJSOverlayView {
  applyingObjects: Map<string, V.Object>;
  deletingObjects: Set<string>;
  objects: Map<string, V.ObjectWrapper>;
  scene: THREE.Scene;

  constructor(position: POS.Position) {
    let coordinate = position.coordinate;
    mapOptions.center = {
      lat: coordinate.latitude,
      lng: coordinate.longitude,
    };

    const viewEl = document.getElementById(mainViewElID) as HTMLElement;
    const map = new google.maps.Map(viewEl, mapOptions);
    const scene = new THREE.Scene();

    const ambientLight = new THREE.AmbientLight(0xffffff, 0.75);
    scene.add(ambientLight);

    const directionalLight = new THREE.DirectionalLight(0xffffff, 0.25);
    directionalLight.position.set(0, 10, 50);
    scene.add(directionalLight);

    super({
      map: map,
      scene: scene,
      anchor: { ...mapOptions.center, altitude: 100 },
    });

    position.addListener((coordinate) => {
      map.setOptions({
        center: {
          lat: coordinate.latitude,
          lng: coordinate.longitude,
        }
      });
    });

    this.applyingObjects = new Map<string, V.Object>();
    this.deletingObjects = new Set<string>();
    this.objects = new Map<string, V.ObjectWrapper>();
    this.scene = scene;
  }

  applyObjects(objects: V.Object[]): void {
    for (let obj of objects) {
      this.applyingObjects.set(obj.meta.uuid, obj);
    }
  }

  deleteObjects(uuids: string[]): void {
    for (let uuid of uuids) {
      this.applyingObjects.delete(uuid);
      this.deletingObjects.add(uuid);
    }
  }

  onDraw({ gl, transformer }: google.maps.WebGLDrawOptions): void {
    super.onDraw({ gl, transformer });

    for (const [uuid, obj] of this.applyingObjects) {
      let wrapper = this.objects.get(uuid);
      if (wrapper === undefined) {
        wrapper = new V.ObjectWrapper(V.ScaleModeLandScape);
        this.objects.set(uuid, wrapper);
        this.scene.add(wrapper);
      }
      wrapper.applyObject(obj);
    }
    this.applyingObjects.clear();

    for (const uuid of this.deletingObjects) {
      let wrapper = this.objects.get(uuid);
      if (wrapper === undefined) {
        continue
      }
      this.scene.remove(wrapper);
      this.objects.delete(uuid);
    }
    this.deletingObjects.clear();

    for (const [_, wrapper] of this.objects) {
      // wrapper.transformPosition(transformer);
      // TODO: this code is workaround. i couldn't find the correct way.
      wrapper.position.copy(this.latLngAltitudeToVector3({
        lat: wrapper.objPosition.y,
        lng: wrapper.objPosition.x,
        altitude: wrapper.objPosition.z,
      }));

    }

    this.requestRedraw();
  }
}
