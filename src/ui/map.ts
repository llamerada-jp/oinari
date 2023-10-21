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

import { AmbientLight, DirectionalLight, Scene } from "three";

import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { ThreeJSOverlayView } from "@googlemaps/three";
import { Keys } from "../keys";


interface ApplyObjectsRequest {
  objects: Object[];
}

interface DeleteObjectsRequest {
  uuids: string[];
}

interface Object {
  meta: ObjectMeta;
  spec: ObjectSpec;
}

interface ObjectMeta {
  type: string;
  name: string;
  owner: string;
  creatorNode: string;
  uuid: string;
  deletionTimestamp: string;
}

interface ObjectSpec {
  parts: PartSpec[];
  materials: MaterialSpec[];
  maps: TextureSpec[];
  position: Vec3;
}

interface PartSpec {
  name: string;
  sprite: SpriteSpec;
}

interface SpriteSpec {
  material: string;
  center: Vec2;
}

interface MaterialSpec {
  name: string;
  color: Color;
  mapTexture: string;
}

interface TextureSpec {
  name: string;
  urlTexture: URLTextureSpec;
}

interface TextureBaseSpec {
}

interface URLTextureSpec extends TextureBaseSpec {
  url: string;
}

interface Color {
  r: number;
  g: number;
  b: number;
}

interface Vec2 {
  x: number;
  y: number;
}

interface Vec3 {
  x: number;
  y: number;
  z: number;
}

let map: google.maps.Map;

const mapOptions = {
  tilt: 0,
  heading: 0,
  zoom: 18,
  center: { lat: 35.6594945, lng: 139.6999859 },
  mapId: Keys.googleMapID,
  // disable interactions due to animation loop and moveCamera
  disableDefaultUI: true,
  gestureHandling: "none",
  keyboardShortcuts: false,
};

export function init(frontendMpx: CL.MultiPlexer): void {
  initHandler(frontendMpx);

  const mapDiv = document.getElementById("map") as HTMLElement;

  map = new google.maps.Map(mapDiv, mapOptions);

  const scene = new Scene();

  const ambientLight = new AmbientLight(0xffffff, 0.75);

  scene.add(ambientLight);

  const directionalLight = new DirectionalLight(0xffffff, 0.25);

  directionalLight.position.set(0, 10, 50);
  scene.add(directionalLight);

  // Load the model.
  const loader = new GLTFLoader();
  const url =
    "https://raw.githubusercontent.com/googlemaps/js-samples/main/assets/pin.gltf";

  loader.load(url, (gltf) => {
    gltf.scene.scale.set(10, 10, 10);
    gltf.scene.rotation.x = Math.PI / 2;
    scene.add(gltf.scene);

    let { tilt, heading, zoom } = mapOptions;

    const animate = () => {
      if (tilt < 67.5) {
        tilt += 0.5;
      } else if (heading <= 360) {
        heading += 0.2;
        zoom -= 0.0005;
      } else {
        // exit animation loop
        return;
      }

      map.moveCamera({ tilt, heading, zoom });

      requestAnimationFrame(animate);
    };

    requestAnimationFrame(animate);
  });

  let overlayView = new ThreeJSOverlayView({
    map,
    scene,
    anchor: { ...mapOptions.center, altitude: 100 },
  });

  overlayView.onDraw = ({gl, transformer}) => {
  };
}

function initHandler(frontendMpx: CL.MultiPlexer): void {
  frontendMpx.setHandlerFunc("applyObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as ApplyObjectsRequest;
    console.log(JSON.stringify(request));
    for (let obj of request.objects) {
      applyObject(obj);
    }
    writer.replySuccess(null);
  });

  frontendMpx.setHandlerFunc("deleteObjects", (data: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    let request = data as DeleteObjectsRequest;
    for (let uuid of request.uuids) {
      deleteObject(uuid);
    }
    writer.replySuccess(null);
  });
}

function applyObject(obj: Object): void {
  console.log("🤖 applyObject:" + JSON.stringify(obj));
}

function deleteObject(uuid: string): void {
  console.log("🤖 deleteObject:" + uuid);
}