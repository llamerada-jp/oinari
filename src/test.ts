import * as CL from "./crosslink";
import * as CRI from "./cri";

import * as crosslinkGoTest from "./crosslink.test.go";

async function testCrossLinkGo() {
  console.log("start test of CrossLink");
  try {
    let crosslinkGoTester = new crosslinkGoTest.Tester()
    console.assert(await crosslinkGoTester.start(), "cross link");
  } catch (e) {
    console.log(e);
  }
}

async function testUsingController() {
  console.log("start tests implements by WASM");
  // start controller worker
  const controller = new Worker("./controller.js");

  // setup crosslink
  let rootMpx = new CL.MultiPlexer();
  let crosslink = new CL.Crosslink(new CL.WorkerImpl(controller), rootMpx);

  // setup CRI
  CRI.initCRI(crosslink, rootMpx);

  // run wasm test program
  await crosslink.call("run", {
    file: "test/test.wasm",
  });
}

async function main() {
  await testCrossLinkGo();
  await testUsingController();
  console.log("FINISH");
}

main();