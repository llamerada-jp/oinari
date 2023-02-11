import * as CL from "./crosslink";
import * as CM from "./command";
import * as WB from "./webrtc_bypass_handler";

declare function ColonioModule(): Promise<any>;

let rootMpx: CL.MultiPlexer;
let crosslink: CL.Crosslink;
let command: CM.Commands;

function initCrosslink() {
  const worker = new Worker("worker.js");
  rootMpx = new CL.MultiPlexer();
  crosslink = new CL.Crosslink(new CL.WorkerImpl(worker), rootMpx);
}

async function initColonio(cl: CL.Crosslink) {
  let mpx = new CL.MultiPlexer();
  rootMpx.setHandler("colonio", mpx);
  let colonio = await ColonioModule();
  let webrtcImpl: WebrtcImplement = new colonio.DefaultWebrtcImplement();
  mpx.setHandler("webrtc", WB.NewWebrtcHandler(cl, webrtcImpl));
}

function initAgent(): Promise<void> {
  return new Promise((resolve) => {
    let mpx = new CL.MultiPlexer();
    rootMpx.setHandler("system", mpx);

    // this handler for waiting to complete initialization of golang module.
    mpx.setRawHandlerFunc("onInitComplete", (_1: string, _2: Map<string, string>, writer: CL.ResponseWriter) => {
      writer.replySuccess("");
      resolve();
    });
  });
}

async function main() {
  initCrosslink();
  await initColonio(crosslink);
  await initAgent();
  command = new CM.Commands(crosslink);
  await command.connect("ws://localhost:8080/seed", "dummy-account", "");
  // set a position for sample playing
  await command.setPosition(35.6594945, 139.6999859);
  // run sample application
  let app = await command.runApplication("./sample.app.yaml");
  setTimeout(() => {
    command.terminateApplication(app.uuid);
  }, 60 * 1000);
}

main();

/*
import { Loader } from '@googlemaps/js-api-loader';
import { AmbientLight, DirectionalLight, Scene } from "three";

import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import { ThreeJSOverlayView } from "@googlemaps/three";
import { Keys } from "./keys";

let map: google.maps.Map;

const apiLoader = new Loader({
  apiKey: Keys.googleApiKey,
  version: "beta",
});

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

apiLoader.load().then((google) => {
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

  new ThreeJSOverlayView({
    map,
    scene,
    anchor: { ...mapOptions.center, altitude: 100 },
  });
});
//*/