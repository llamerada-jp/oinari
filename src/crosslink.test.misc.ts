import * as CL from "./crosslink";

export class workerMock implements CL.WorkerInterface {
  pair!: workerMock;
  listener!: (datum: any) => void;

  addEventListener(listener: (datum: any) => void): void {
    this.listener = listener;
  }

  post(datum: object): void {
    this.pair.listener(datum);
  }

  setPair(pair: workerMock) {
    this.pair = pair;
  }
}

export function makeTags(path: string): Map<string, string> {
  let r = new Map<string, string>();
  r.set(CL.TAG_PATH, path);
  return r;

}
