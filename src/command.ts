import * as CL from "./crosslink";

const CL_SYSTEM_PATH: string = "system";
const CL_MANAGER_PATH: string = "podManager";

export interface ApplicationDigest {
  name: string
  uuid: string
  runningNode: string
  owner: string
}

interface ApplicationDefinition {
  metadata: ApplicationMetadata
  spec: PodSpec
}

interface ApplicationMetadata {
  name: string
}

interface ManagerRunRequest {
  name: string
  spec: PodSpec
}

interface ManagerRunResponse {
  digest: ApplicationDigest
}

interface ManagerListResponse {
  digests: Array<ApplicationDigest>
}

interface ManagerTerminateRequest {
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
}

interface Container {
  name: string
  image: string
}

export class Commands {
  private cl: CL.Crosslink;
  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  connect(url: string, account: string, token: string): Promise<any> {
    return this.cl.call(CL_SYSTEM_PATH + "/connect", {
      url: url,
      account: account,
      token: token,
    });
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
      let app = a as ApplicationDefinition;
      let name = app.metadata.name;
      if (postfix != null) {
        name = name + postfix;
      }
      return this.cl.call(CL_MANAGER_PATH + "/run", {
        name: name,
        spec: app.spec
      } as ManagerRunRequest);

    }).then((r) => {
      let response = r as ManagerRunResponse;
      return response.digest;
    });
  }

  listApplications(): Promise<Array<ApplicationDigest>> {
    return this.cl.call(CL_MANAGER_PATH + "/list", {}).then((r) => {
      let response = r as ManagerListResponse;
      return response.digests;
    });
  }

  terminateApplication(uuid: string): Promise<any> {
    return this.cl.call(CL_MANAGER_PATH + "/terminate", {
      uuid: uuid,
    } as ManagerTerminateRequest);
  }
}