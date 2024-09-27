import { GraphEdge } from "reagraph";
import NodeService, { Node, Properties } from "./nodeService.ts";
import apiClient from "./api-client.ts";

export interface Relationship {
  ID: number;
  StartID: number;
  EndID: number;
  Kind: string;
  Properties: Properties;
}

export enum relationships {
  ActsOn = "ActsOn",
  AllowAction = "AllowAction",
  AttachedTo = "AttachedTo",
  DenyAction = "DenyAction",
  ExpandsTo = "ExpandsTo",
  InAccount = "InAccount",
  MemberOf = "MemberOf",
  OnResource = "OnResource",
  TypeOf = "TypeOf",
}

export function relationshipToGraphEdge(relationship: Relationship): GraphEdge {
  const size = relationship.Properties.map.layer.toString() == "2" ? 7 : 3;
  return {
    id: relationship.ID.toString(),
    source: relationship.StartID.toString(),
    target: relationship.EndID.toString(),
    label: relationship.Kind === "IdentityTransform" ? relationship.Properties.map.name : relationship.Kind,
    size: size,
  };
}

export function getRelationshipByID(relationshipId: string) {
  const controller = new AbortController();

  const request = apiClient.get<Relationship>(
    `/relationship/${relationshipId}`,
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

export async function GetRelationshipNodes(
  relationship: Relationship
): Promise<[Node, Node]> {
  const { request: startNode } = NodeService.getNodeByID(
    relationship.StartID.toString()
  );
  const { request: endNode } = NodeService.getNodeByID(
    relationship.EndID.toString()
  );

  try {
    const results = await Promise.all([
      (await startNode).data,
      (await endNode).data,
    ]);
    return results;
  } catch (error) {
    throw error;
  }
}
