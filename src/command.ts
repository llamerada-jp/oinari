import * as CL from "./crosslink";

export class Commands {
  private cl: CL.Crosslink;
  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  connect(url: string, token: string): Promise<any> {
    return this.cl.call("system/connect", {
      url: url,
      token: token,
    });
  }

  applyPod(name: string, image: string): Promise<{ uuid: string }> {
    return this.cl.call("system/applyPod", {
      name: name,
      image: image,
    }).then((result) => {
      return {
        uuid: (result as any).uuid,
      };
    });
  }

  setPosition(lat: number, lon: number): Promise<any> {
    return this.cl.call("system/setPosition", {
      latitude: lat,
      longitude: lon,
    });
  }

  terminate(uuid: string): Promise<any> {
    return this.cl.call("system/terminate", {
      uuid: uuid,
    });
  }
}