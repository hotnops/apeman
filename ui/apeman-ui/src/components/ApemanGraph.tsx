import { useRef } from "react";

import {
  GraphCanvas,
  GraphCanvasRef,
  InternalGraphEdge,
  InternalGraphNode,
} from "reagraph";
import nodeService, { Node } from "../services/nodeService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { getRelationshipByID } from "../services/relationshipServices";

const theme = {
  canvas: {
    background: "#f2f2f0",
  },
  node: {
    fill: "#7CA0AB",
    activeFill: "#1DE9AC",
    opacity: 1,
    selectedOpacity: 1,
    inactiveOpacity: 0.2,
    label: {
      color: "#2A6475",
      stroke: "#f2f2f0",
      activeColor: "#1DE9AC",
    },
    subLabel: {
      color: "#ddd",
      stroke: "transparent",
      activeColor: "#1DE9AC",
    },
  },
  lasso: {
    border: "1px solid #55aaff",
    background: "rgba(75, 160, 255, 0.1)",
  },
  ring: {
    fill: "#D8E6EA",
    activeFill: "#1DE9AC",
  },
  edge: {
    fill: "#D8E6EA",
    activeFill: "#1DE9AC",
    opacity: 1,
    selectedOpacity: 1,
    inactiveOpacity: 0.1,
    label: {
      stroke: "#f2f2f0",
      color: "#2A6475",
      activeColor: "#1DE9AC",
      fontSize: 6,
    },
  },
  arrow: {
    fill: "#D8E6EA",
    activeFill: "#1DE9AC",
  },
  cluster: {
    stroke: "#D8E6EA",
    opacity: 1,
    selectedOpacity: 1,
    inactiveOpacity: 0.1,
    label: {
      stroke: "#f2f2f0",
      color: "#2A6475",
    },
  },
};

interface Props {
  setGraphNodes: React.Dispatch<React.SetStateAction<{ [key: string]: Node }>>;
}

const ApemanGraph = ({ setGraphNodes }: Props) => {
  const graphRef = useRef<GraphCanvasRef | null>(null);
  const { nodes, edges, activeElement, setActiveElement } = useApemanGraph();
  const divRef = useRef<HTMLDivElement>(null);

  var activeElementId = null;

  if (activeElement) {
    activeElementId =
      "id" in activeElement ? activeElement.id : activeElement.ID;
  }

  const canvasOptions = {
    antialias: true,
    preserveDrawingBuffer: false,
    depth: true,
    stencil: false,
    alpha: true,
    premultipliedAlpha: true,
    failIfMajorPerformanceCaveat: true,
    // powerPreference: "high-performance",
    desynchronized: false,
  };

  return (
    <div ref={divRef}>
      <GraphCanvas
        ref={graphRef}
        nodes={nodes}
        edges={edges}
        glOptions={canvasOptions}
        edgeLabelPosition="inline"
        edgeInterpolation="linear"
        labelType="all"
        theme={theme}
        layoutType="forceDirected2d"
        selections={activeElementId ? [activeElementId.toString()] : []}
        onCanvasClick={() => {
          setActiveElement(null);
        }}
        onNodeClick={(n: InternalGraphNode) => {
          nodeService.getNodeByID(n.id).request.then((res) => {
            setActiveElement(res.data);
            setGraphNodes((prevGraphNodes) => {
              const newNodes = { ...prevGraphNodes };
              newNodes[res.data.id] = res.data;
              return newNodes;
            });
          });
        }}
        onEdgeClick={(e: InternalGraphEdge) => {
          getRelationshipByID(e.id).request.then((res) => {
            if (res.data.Properties.map.layer.toString() == "2") {
              setActiveElement(res.data);
            }
          });
        }}
        draggable
      />
    </div>
  );
};

export default ApemanGraph;
