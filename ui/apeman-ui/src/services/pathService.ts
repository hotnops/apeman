import apiClient from "./api-client";
import { Node } from "./nodeService";
import { Relationship } from "./relationshipServices";

export interface Path {
  Nodes: Node[];
  Edges: Relationship[];
}

export interface ActionPathEntry {
  principal_id: number;
  principal_arn: string;
  resource_arn: string;
  action: string;
  path: Path;
  effect: string;
  statement: Node;
  conditions: Node[];
}

export interface ActionPathSet {
  action_paths: ActionPathEntry[];
}

export interface PathResponse {
  paths: Path[];
}

export function getNodesFromPath(path: Path) {
  return path.Nodes;
}

export function getNodesFromPaths(paths: Path[]) {
  var nodes: Node[] = [];
  paths.map((path) => {
    path.Nodes.map((node) => {
      nodes = nodes.filter((oldNode) => node.id != oldNode.id);
      nodes.push(node);
    });
  });
  return nodes;
}

export function addPathToGraph(
  path: Path,
  addNode: (n: Node) => void,
  addEdge: (e: Relationship) => void
) {
  path.Nodes.map((node) => {
    addNode(node);
  });

  path.Edges.map((edge) => {
    addEdge(edge);
  });
}

export function GetNodeShortestPath(startNodeId: number, endNodeId: number) {
  const controller = new AbortController();
  const request = apiClient.get(
    `/node/${startNodeId}/shortestpath/${endNodeId}`,
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

export function GetNodeIdentityPath(startNodeId: number, endNodeId: number) {
  const controller = new AbortController();
  const request = apiClient.get(
    `/node/${startNodeId}/identitypath/${endNodeId}`,
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

export function GetNodePermissionPath(startNodeId: number, endNodeId: number) {
  const controller = new AbortController();
  const request = apiClient.get(
    `/node/${startNodeId}/permissionpath/${endNodeId}`,
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

export function GetNodePermissionPathWithAction(
  startNodeId: number,
  endNodeId: number,
  action: string
) {
  const controller = new AbortController();
  const request = apiClient.get<Path[]>(
    `/node/${startNodeId}/permissionpath/${endNodeId}?action=${action}`,
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
