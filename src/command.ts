import * as CL from "./crosslink";

const CL_SYSTEM_PATH: string = "node/system";
const CL_RESOURCE_PATH: string = "node/resource";

export const CONFIG_KEY_ALLOW_APPLICATIONS: string = "allowApplications";
export const CONFIG_KEY_SAMPLE_PREFIX: string = "samplePrefix";

export interface NodeInfo {
  commitHash: string
}

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
  position: Vector3 | null
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

interface Vector3 {
  x: number
  y: number
  z: number
}

export class Commands {
  private cl: CL.Crosslink;
  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  getNodeInfo(): Promise<NodeInfo> {
    return this.cl.call(CL_SYSTEM_PATH + "/info", {}) as Promise<NodeInfo>;
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

  setConfiguration(key: string, value: string): void {
    this.cl.call(CL_SYSTEM_PATH + "/config", {
      key: key,
      value: value,
    } as Config);
  }

  // position.x: latitude[degree]
  // position.y: longitude[degree]
  // position.z: altitude[meter]
  setPosition(position: Vector3): Promise<any> {
    return this.cl.call(CL_RESOURCE_PATH + "/setNodePosition", {
      position: position,
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

  runApplication(url: string): Promise<ApplicationDigest> {
    let app: ApplicationDefinition;
    return fetch(url).then((response: Response) => {
      if (!response.ok) {
        throw new Error("Application could not start: " + response.statusText);
      }
      return response.json();

    }).then((a) => {
      app = a as ApplicationDefinition

      return this.cl.call(CL_RESOURCE_PATH + "/listPod", {});

    }).then((l) => {
      let pods = l as ListPodResponse;
      let podNames = pods.digests.map((d) => d.name);

      for (let container of app.spec.containers) {
        container.image = new URL(container.image, url).toString();
      }
      let name = app.metadata.name;

      // put a number suffix if the name is already used.
      if (podNames.includes(name)) {
        let i = 1;
        while (podNames.includes(name + "-" + i)) {
          i++;
        }
        name = name + "-" + i;
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
}