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

import * as AFRAME from "aframe"

const mainViewElID = "mainView";

let xrView: XrView;

export function init(frontendMpx: CL.MultiPlexer, pos: POS.Position): void {
  xrView = new XrView(pos);
  document.body.style.overflow = "hidden";
  initHandler(frontendMpx);
}

function initHandler(frontendMpx: CL.MultiPlexer): void {
  frontendMpx.setHandlerFunc("applyObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as V.ApplyObjectsRequest;
    xrView.applyObjects(request.objects);
    writer.replySuccess(null);
  });

  frontendMpx.setHandlerFunc("deleteObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as V.DeleteObjectsRequest;
    xrView.deleteObjects(request.uuids);
    writer.replySuccess(null);
  });
}

class XrView {
  scene: AFRAME.Scene;
  camera: AFRAME.Entity;
  objects: Map<string, V.ObjectWrapper>;
  entities: Map<string, AFRAME.Entity>;

  constructor(pos: POS.Position) {
    let viewEl = document.getElementById(mainViewElID);
    viewEl!.classList.remove("d-none");

    this.scene = document.createElement("a-scene") as AFRAME.Scene;
    /* Added sourceWidth and sourceHeight parameter to suppress body size rewriting by AR.js.
     * This change makes the MDB UI display correctly, but the size of the video element has
     * changed from the original AR.js created.
     * Therefore, the size of the AR space may not be correct. */
    let sourceWidth = window.innerWidth;
    let sourceHeight = window.innerHeight;
    this.scene.setAttribute("arjs", `trackingMethod: best; sourceType: webcam; debugUIEnabled: false; sourceWidth: ${sourceWidth}; sourceHeight: ${sourceHeight};`);
    this.scene.setAttribute("embedded", "");
    this.scene.setAttribute("renderer", "logarithmicDepthBuffer: true;");
    this.scene.setAttribute("vr-mode-ui", "enabled: false;");
    viewEl?.appendChild(this.scene);

    this.camera = document.createElement("a-camera") as AFRAME.Entity;
    let cameraAttr = "";
    if (!pos.enableGNSS) {
      let latitude = pos.coordinate.latitude;
      let longitude = pos.coordinate.longitude;
      let altitude = pos.coordinate.altitude;
      cameraAttr += `simulateLatitude:${latitude}; simulateLongitude:${longitude}; simulateAltitude:${altitude};`;
    }
    this.camera.setAttribute("gps-camera", cameraAttr);
    this.camera.setAttribute("rotation-reader", "");
    this.scene.appendChild(this.camera);

    this.objects = new Map<string, V.ObjectWrapper>();
    this.entities = new Map<string, AFRAME.Entity>();
  }

  applyObjects(objects: V.Object[]): void {
    for (let obj of objects) {
      let uuid = obj.meta.uuid;
      let entity = this.entities.get(uuid);
      let wrapper = this.objects.get(uuid);

      if (entity === undefined || wrapper === undefined) {
        entity = document.createElement("a-entity");
        this.entities.set(uuid, entity);
        entity.setAttribute("scale", "1 1 1");

        wrapper = new V.ObjectWrapper(V.ScaleModeXR);
        this.objects.set(uuid, wrapper);

        entity.object3D.add(wrapper);
        this.scene.appendChild(entity);
      }

      wrapper.applyObject(obj);

      let latitude = obj.spec.position.y;
      let longitude = obj.spec.position.x;
      // let altitude = obj.spec.position.z;
      // TODO apply altitude
      entity.setAttribute("gps-entity-place", `latitude:${latitude}; longitude:${longitude};`);
    }
  }

  deleteObjects(uuids: string[]): void {
    for (let uuid of uuids) {
      let entity = this.entities.get(uuid);
      let wrapper = this.objects.get(uuid);
      if (entity === undefined || wrapper === undefined) {
        continue;
      }

      entity.object3D.remove(wrapper);
      this.scene.removeChild(entity);
    }
  }
}