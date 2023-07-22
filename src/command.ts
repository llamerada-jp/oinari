import * as CL from "./crosslink";

const CL_SYSTEM_PATH: string = "system";
const CL_RESOURCE_PATH: string = "resource";

export interface ConnectInfo {
  account: string
  node: string
}

export interface ApplicationDigest {
  name: string
  uuid: string
  runningNode: string
  owner: string
  phase: string
}

interface ApplicationDefinition {
  metadata: ApplicationMetadata
  spec: PodSpec
}

interface ApplicationMetadata {
  name: string
}

interface CreatePodRequest {
  name: string
  spec: PodSpec
}

interface CreatePodResponse {
  digest: ApplicationDigest
}

interface ListPodResponse {
  digests: Array<ApplicationDigest>
}

interface DeletePodRequest {
  uuid: string
}

interface ObjectMeta {
  name: string
  namespace: string | undefined
}

interface Pod {
  meta: ObjectMeta
  spec: PodSpec
}

interface PodSpec {
  containers: Array<Container>
  // TODO scheduler
}

interface Container {
  name: string
  image: string
  runtime: Array<string>
  args: Array<string>
  env: Array<EnvVar>
}

interface EnvVar {
  name: string
  value: string
  // valueFrom is not supported yet.
}

export class Commands {
  private cl: CL.Crosslink;
  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  connect(url: string, account: string, token: string): Promise<ConnectInfo> {
    return this.cl.call(CL_SYSTEM_PATH + "/connect", {
      url: url,
      account: account,
      token: token,
    }) as Promise<ConnectInfo>;
  }

  setPosition(lat: number, lon: number): Promise<any> {
    return this.cl.call(CL_SYSTEM_PATH + "/setPosition", {
      latitude: lat,
      longitude: lon,
    });
  }

  runApplication(url: string, postfix?: string | undefined): Promise<ApplicationDigest> {
    return fetch(url).then((response: Response) => {
      if (!response.ok) {
        throw new Error("Application could not start: " + response.statusText);
      }
      return response.json();

    }).then((a) => {
      let app = a as ApplicationDefinition
      for (let container of app.spec.containers) {
        container.image = new URL(container.image, url).toString();
      }
      let name = app.metadata.name;
      if (postfix != null) {
        name = name + postfix;
      }
      return this.cl.call(CL_RESOURCE_PATH + "/createPod", {
        name: name,
        spec: app.spec
      } as CreatePodRequest);

    }).then((r) => {
      let response = r as CreatePodResponse;
      return response.digest;
    });
  }

  listProcess(): Promise<Array<ApplicationDigest>> {
    return this.cl.call(CL_RESOURCE_PATH + "/listPod", {}).then((r) => {
      let response = r as ListPodResponse;
      return response.digests;
    });
  }

  terminateProcess(uuid: string): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/deletePod", {
      uuid: uuid,
    } as DeletePodRequest);
  }
}