import { GraphNode } from "reagraph";
import apiClient from "./api-client";
import { Path } from "./pathService";

type PropertyMap = {
  [key: string]: string;
};

export interface Properties {
  map: PropertyMap;
  deleted: {};
  modified: {};
}

export interface Node {
  id: number;
  kinds: string[];
  properties: Properties;
}

export interface NodeResponse {
  version: string
  count: number
  nodes: Node[]
}

export enum kinds {
  AWSAccount = "AWSAccount",
  AWSRole = "AWSRole",
  AWSUser = "AWSUser",
  AWSGroup = "AWSGroup",
  AWSManagedPolicy = "AWSManagedPolicy",
  AWSInlinePolicy = "AWSInlinePolicy",
  AWSPolicyDocument = "AWSPolicyDocument",
  AWSStatement = "AWSStatement",
  AWSAssumeRolePolicy = "AWSAssumeRolePolicy",
  AWSPolicyVersion = "AWSPolicyVersion",
  AWSResourceType = "AWSResourceType",
  AWSResourceBlob = "AWSResourceBlob",
  AWSActionBlob = "AWSActionBlob",
  AWSAction = "AWSAction",
  AWSCondition = "AWSCondition",
  AWSConditionKey = "AWSConditionKey",
  AWSConditionValue = "AWSConditionValue",
  AWSConditionOperator = "AWSConditionOperator",
  UniqueArn = "UniqueArn",
  UniqueName = "UniqueName",
}

const NODE_BASE = "/nodes";

export function getIconURL(nodeKinds: string[]): string {
  const icon_dir = "./";
  if (nodeKinds.includes(kinds.AWSAccount)) {
    return icon_dir + "account_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSRole)) {
    return icon_dir + "role_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSUser)) {
    return icon_dir + "user_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSGroup)) {
    return icon_dir + "group_icon.svg";
  }
  if (
    nodeKinds.includes(kinds.AWSManagedPolicy) ||
    nodeKinds.includes(kinds.AWSInlinePolicy)
  ) {
    return icon_dir + "policy_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSPolicyDocument)) {
    return icon_dir + "policy_document_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSStatement)) {
    return icon_dir + "statement_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSAssumeRolePolicy)) {
    return icon_dir + "assume_role_policy_icon.svg";
  }
  // if (kinds.includes("AWSPolicyVersion")) {
  //   return icon_dir + "/statement_icon.svg"
  // }
  if (nodeKinds.includes(kinds.AWSResourceType)) {
    return icon_dir + "resource_type_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSResourceBlob)) {
    return icon_dir + "resource_blob_icon.svg";
  }
  // if (kinds.includes("AWSAction")) {
  //   return icon_dir + "/statement_icon.svg"
  // }
  if (nodeKinds.includes(kinds.AWSActionBlob)) {
    return icon_dir + "action_blob_icon.svg";
  }
  if (nodeKinds.includes(kinds.AWSCondition)) {
    return icon_dir + "condition_icon.svg";
  }
  // if (kinds.includes("AWSConditionKey")) {
  //   return icon_dir + "/statement_icon.svg"
  // }
  // if (kinds.includes("AWSConditionValue")) {
  //   return icon_dir + "/statement_icon.svg"
  // }
  // if (kinds.includes("AWSConditionOperator")) {
  //   return icon_dir + "/condition_operator_icon.svg"
  // }
  // if (kinds.includes("AWSMultivalueOperator")) {
  //   return icon_dir + "/condition_operator_icon.svg"
  // }

  if (nodeKinds.includes(kinds.UniqueArn)) {
    return icon_dir + "resource_icon.svg";
  }
  return "";
}

export function getNodeLabel(node: Node): string {
  const nodeKinds = node.kinds;
  if (nodeKinds.includes(kinds.AWSAccount)) {
    return node.properties.map.account_id;
  }
  if (nodeKinds.includes(kinds.AWSRole)) {
    return node.properties.map.rolename;
  }
  if (nodeKinds.includes(kinds.AWSUser)) {
    return node.properties.map.name;
  }
  if (nodeKinds.includes(kinds.AWSGroup)) {
    return node.properties.map.name;
  }
  if (nodeKinds.includes(kinds.AWSManagedPolicy)) {
    return node.properties.map.policyname;
  }
  if (nodeKinds.includes(kinds.UniqueArn)) {
    return node.properties.map.arn;
  }
  if (nodeKinds.includes(kinds.UniqueName)) {
    return node.properties.map.name;
  }
  if (nodeKinds.includes(kinds.AWSStatement)) {
    return node.properties.map.sid
      ? node.properties.map.sid
      : node.id.toString();
  }
  if (nodeKinds.includes(kinds.AWSPolicyVersion)) {
    return node.properties.map.versionid;
  }
  return node.id.toString();
}

function getNodeFill(node: Node) {
  if (node.kinds.includes(kinds.AWSStatement)) {
    if (node.properties.map.effect == "Allow") {
      return "#76d654";
    } else {
      return "#de6e68";
    }
  }
  return undefined;
}

export function nodeToGraphNode(node: Node): GraphNode {
  return {
    id: node.id.toString(),
    label: getNodeLabel(node),
    icon: getIconURL(node.kinds),
    fill: getNodeFill(node),
  };
}

class NodeService {
  getNodeByID(nodeId: string) {
    const controller = new AbortController();

    const request = apiClient.get<Node>(NODE_BASE + "/" + nodeId, {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getAllNodes() {
    const controller = new AbortController();

    const request = apiClient.get<Node[]>(NODE_BASE, {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getTierZeroNodes(account_id: string) {
    const controller = new AbortController();

    const request = apiClient.get<Node[]>(
      `${NODE_BASE}/${account_id}/tierzero`,
      {
        signal: controller.signal,
      }
    );
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getTierZeroPaths(account_id: string) {
    const controller = new AbortController();

    const request = apiClient.get<Path[]>(
      `${NODE_BASE}/${account_id}/tierzeropaths`,
      {
        signal: controller.signal,
      }
    );
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getNodeTags(nodeId: string) {
    const controller = new AbortController();

    const request = apiClient.get<Node[]>(NODE_BASE + "/" + nodeId + "/tags", {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getNodesWithParams<T>(params: Map<string, string>) {
    const controller = new AbortController();

    var paramArray: Array<string> = [];

    Array.from(params).map(([key, value]) => {
      paramArray.push(key + "=" + value);
    });

    const queryString = paramArray.join("&");

    const request = apiClient.get<T[]>(NODE_BASE + "?" + queryString, {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getNodesByKind<T>(kind: string) {
    const controller = new AbortController();

    const request = apiClient.get<T[]>(NODE_BASE + "?kind=" + kind, {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getAttachedNodes(
    nodeId: string,
    direction: string,
    relkinds: string[] = [],
    kinds: string[] = []
  ) {
    const controller = new AbortController();

    var path = NODE_BASE + "/" + nodeId + "/";
    if (direction === "outbound") {
      path += "outboundedges?";
    } else if (direction === "inbound") {
      path += "inboundedges?";
    }

    const queryParams = new URLSearchParams();
    relkinds.map((rel) => queryParams.append("relkind", rel));
    kinds.map((kind) => queryParams.append("kind", kind));
    path += queryParams.toString();

    const request = apiClient.get(path);

    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getNodeByArn(arn: string) {
    const controller = new AbortController();

    const request = apiClient.get<Node[]>(NODE_BASE + "?arn=" + arn, {
      signal: controller.signal,
    });
    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }
}

export default new NodeService();
