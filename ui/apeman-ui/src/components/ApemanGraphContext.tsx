import { createContext, useState, ReactNode } from "react";
import { GraphNode, GraphEdge } from "reagraph";
import { Node, nodeToGraphNode } from "../services/nodeService";
import {
  Relationship,
  relationships,
  relationshipToGraphEdge,
} from "../services/relationshipServices";

interface ApemanGraphContextType {
  activeElement: Node | Relationship | null;
  nodes: GraphNode[];
  edges: GraphEdge[];
  selections: string[];
  addNode: (node: Node) => void;
  addEdge: (edge: Relationship) => void;
  setNodes: (nodes: GraphNode[]) => void;
  setEdges: (edges: GraphEdge[]) => void;
  setActiveElement: (n: Node | Relationship | null) => void;
}

export const ApemanGraphContext = createContext<
  ApemanGraphContextType | undefined
>(undefined);

export const ApemanGraphProvider = ({ children }: { children: ReactNode }) => {
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [edges, setEdges] = useState<GraphEdge[]>([]);
  const [selections, setSelections] = useState<string[]>([]);
  const [activeElement, setActiveElementInternal] = useState<
    Node | Relationship | null
  >(null);

  const addNode = (node: Node) => {
    const graphNode = nodeToGraphNode(node);
    const index = nodes.findIndex((item) => item.id === graphNode.id);
    if (index === -1) {
      console.log(graphNode);
      setNodes((prevNodes) => [...prevNodes, graphNode]);
    }
  };
  const addEdge = (edge: Relationship) => {
    const graphEdge = relationshipToGraphEdge(edge);
    setEdges((prevEdges) => [...prevEdges, graphEdge]);
  };

  function setActiveElement(node: Node | Relationship | null) {
    // TODO: fix rel shape to have lowercase id
    if (node == null) {
      setSelections([]);
      setActiveElementInternal(null);
      return;
    }
    const id = "id" in node ? node.id : node.ID;

    setSelections([id.toString()]);
    setActiveElementInternal(node);
  }

  return (
    <ApemanGraphContext.Provider
      value={{
        nodes,
        edges,
        selections,
        activeElement: activeElement,
        addNode,
        addEdge,
        setNodes,
        setEdges,
        setActiveElement: setActiveElement,
      }}
    >
      {children}
    </ApemanGraphContext.Provider>
  );
};
