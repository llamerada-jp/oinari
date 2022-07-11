import * as crosslinkGoTest from "./crosslink.test.go";

async function main() {
  try {
    let crosslinkGoTester = new crosslinkGoTest.Tester()
    console.assert(await crosslinkGoTester.start(), "cross link");
  } catch (e) {
    console.log(e);
  }
}

main();