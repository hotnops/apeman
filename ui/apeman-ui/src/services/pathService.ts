import { useApemanGraph } from "../hooks/useApemanGraph";
import apiClient from "./api-client";
import nodeService, { Node } from "./nodeService";
import { Relationship } from "./relationshipServices";

export interface Path {
  Nodes: Node[];
  Edges: Relationship[];
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
    console.log("EDGE");
    console.log(edge);
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
