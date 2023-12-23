import * as CL from "./crosslink";
import * as CM from "./command";
import * as WB from "./webrtc_bypass_handler";
import * as CRI from "./cri";

import * as crosslinkGoTest from "./crosslink.test.go";

declare function ColonioModule(): Promise<any>;

const TESTS = [
  "test/test_api_core.wasm",
  "test/test_node.wasm",
];

async function testCrossLinkGo() {
  console.log("testing CrossLink");
  try {
    let crosslinkGoTester = new crosslinkGoTest.Tester()
    console.assert(await crosslinkGoTester.start(), "cross link");
  } catch (e) {
    console.log(e);
  }
}

async function testUsingController() {
  // start controller worker
  const controller = new Worker("./controller.js");

  // setup crosslink
  let rootMpx = new CL.MultiPlexer();
  let crosslink = new CL.Crosslink(new CL.WorkerImpl(controller), rootMpx);

  // setup CRI
  CRI.initCRI(crosslink, rootMpx);

  // setup colonio module handler
  let colonioMpx = new CL.MultiPlexer();
  rootMpx.setHandler("colonio", colonioMpx);
  let colonio = await ColonioModule();
  let webrtcImpl: WebrtcImplement = new colonio.DefaultWebrtcImplement();
  colonioMpx.setHandler("webrtc", WB.NewWebrtcHandler(crosslink, webrtcImpl));

  // run wasm test programs
  for (const file of TESTS) {
    console.log("testing " + file);
    await crosslink.call("run", {
      file: file,
    });
  }
}

async function test() {
  console.log("TEST START");
  await testCrossLinkGo();
  await testUsingController();
  console.log("SUCCESS");
}

interface SideNodeParam {
  account: string
  nodeName: string
  nodeType: string
  latitude: number
  longitude: number
  altitude: number
}

const SIDE_NODE_PARAMS: Record<string, SideNodeParam> = {
  "0": {
    account: "cat",
    nodeName: "cat home",
    nodeType: "PC",
    latitude: 45,
    longitude: 100,
    altitude: 0,
  },
  "1": {
    account: "dog",
    nodeName: "dog home",
    nodeType: "PC",
    latitude: 45,
    longitude: 101,
    altitude: 1,
  }
};

async function sidenode(param: SideNodeParam) {
  console.log("RUN SIDENODE");


  // start controller worker
  const controller = new Worker("controller.js");

  // setup crosslink
  let rootMpx = new CL.MultiPlexer();
  let frontendMpx = new CL.MultiPlexer();
  rootMpx.setHandler("frontend", frontendMpx);
  let crosslink = new CL.Crosslink(new CL.WorkerImpl(controller), rootMpx);
  // setup CRI
  CRI.initCRI(crosslink, rootMpx);

  // setup colonio module handler
  let colonio = await ColonioModule();
  let webrtcImpl: WebrtcImplement = new colonio.DefaultWebrtcImplement();
  frontendMpx.setHandler("colonio", WB.NewWebrtcHandler(crosslink, webrtcImpl));

  // setup system module handler
  let systemMpx = new CL.MultiPlexer();
  rootMpx.setHandler("system", systemMpx);

  systemMpx.setHandlerFunc("nodeReady", (_1: any, _2: Map<string, string>, writer: CL.ResponseWriter) => {
    writer.replySuccess("");
    let command = new CM.Commands(crosslink);
    command.connect("https://localhost:8080/seed", param.account, "", param.nodeName, param.nodeType).then(() => {
      return command.setPosition({
        x: param.longitude,
        y: param.latitude,
        z: param.altitude,
      });

    }).then(() => {
      console.log("STANDBY");

    }).catch((e) => {
      console.error(e);
    });
  });

  // run wasm program of controller
  crosslink.call("run", {
    file: "oinari.wasm",
  });
}

async function main() {
  // run as a side node program when URL parameter contains `sidenode` keyword.
  let url = new URL(window.location.href);
  let params = url.searchParams;
  let p = params.get("sidenode");
  if (p != null) {
    let snp = SIDE_NODE_PARAMS[p];

    if (snp == null) {
      console.log("unsupported sidenode index");
      return;
    }
    await sidenode(snp);

  } else {
    // run test other wise.
    await test();
  }
}

main();