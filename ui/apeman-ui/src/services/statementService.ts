import nodeService, { Node } from "./nodeService";
import {
  GetRelationshipNodes,
  Relationship,
  relationships,
} from "./relationshipServices";

export interface StatementDetails {
  policies: Node[];
  actions: Node[];
  resources: Node[];
  conditions: Node[];
}

export async function fetchAllStatementData(
  node: Node,
  setStatementDetails: (s: StatementDetails) => void
) {
  async function fetchRelationshipNodes(
    relationshipType: string[],
    direction: string
  ) {
    const relationships = await nodeService
      .getAttachedNodes(node.id.toString(), direction, relationshipType)
      .request.then((res) => res.data.relationships);
    return Promise.all(
      relationships.map((rel: Relationship) => GetRelationshipNodes(rel))
    );
  }

  try {
    const [actionNodes, resourceNodes, conditionNodes] = await Promise.all([
      fetchRelationshipNodes(
        [relationships.AllowAction, relationships.DenyAction],
        "outbound"
      ),
      fetchRelationshipNodes([relationships.OnResource], "outbound"),
      fetchRelationshipNodes([relationships.AttachedTo], "inbound"),
    ]);

    const mapNodes = (results: any) => results.map((nodes: Node[]) => nodes[1]);

    setStatementDetails({
      actions: mapNodes(actionNodes),
      resources: mapNodes(resourceNodes),
      conditions: mapNodes(conditionNodes),
      policies: [],
    });
  } catch (error) {
    console.error("Failed to fetch statement data:", error);
    // Handle the error appropriately
  }
}
