import * as CL from "./crosslink";

const CL_SYSTEM_PATH: string = "node/system";
const CL_RESOURCE_PATH: string = "node/resource";

export const CONFIG_KEY_ALLOW_APPLICATIONS: string = "allowApplications";
export const CONFIG_KEY_SAMPLE_PREFIX: string = "samplePrefix";

export interface ConnectInfo {
  account: string
  nodeID: string
}

export interface ApplicationDigest {
  name: string
  uuid: string
  runningNodeID: string
  owner: string
  state: string
}

export interface NodeState {
  name: string
  id: string
  account: string
  nodeType: string
  latitude: number | undefined
  longitude: number | undefined
  altitude: number | undefined
}

interface ListNodeResponse {
  nodes: Array<NodeState>
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

interface MigratePodRequest {
  uuid: string
  targetNode: string
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
  enableMigrate: boolean
}

interface Container {
  name: string
  image: string
  runtime: Array<string>
  args: Array<string>
  env: Array<EnvVar>
  restartPolicy: string
}

interface EnvVar {
  name: string
  value: string
  // valueFrom is not supported yet.
}

interface Config {
  key: string
  value: string
}

export class Commands {
  private cl: CL.Crosslink;
  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  connect(url: string, account: string, token: string, nodeName: string, nodeType: string): Promise<ConnectInfo> {
    return this.cl.call(CL_SYSTEM_PATH + "/connect", {
      url: url,
      account: account,
      token: token,
      nodeName: nodeName,
      nodeType: nodeType,
    }) as Promise<ConnectInfo>;
  }

  disconnect(): Promise<any> {
    return this.cl.call(CL_SYSTEM_PATH + "/disconnect", {});
  }

  // lat: latitude[degree]
  // lon: longitude[degree]
  // alt: altitude[meter]
  setPosition(lat: number, lon: number, alt: number): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/setNodePosition", {
      latitude: lat,
      longitude: lon,
      altitude: alt,
    });
  }

  setPublicity(range: number): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/setNodePublicity", {
      range: range,
    });
  }

  listNode(): Promise<Array<NodeState>> {
    return this.cl.call(CL_RESOURCE_PATH + "/listNode", {}).then((r) => {
      let response = r as ListNodeResponse;
      return response.nodes;
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

  migrateProcess(uuid: string, targetNode: string): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/migratePod", {
      uuid: uuid,
      targetNode: targetNode,
    } as MigratePodRequest)
  }

  terminateProcess(uuid: string): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/deletePod", {
      uuid: uuid,
    } as DeletePodRequest);
  }

  setConfiguration(key: string, value: string): void {
    this.cl.call(CL_RESOURCE_PATH + "/config", {
      key: key,
      value: value,
    } as Config);
  }
}