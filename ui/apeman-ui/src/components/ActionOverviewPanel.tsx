<<<<<<< HEAD
import React from "react";
import { Node } from "../services/nodeService";
=======
import { Accordion, Box } from "@chakra-ui/react";
import PathAccordionList from "./PathAccordionList";
import { addPathToGraph, Path } from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { useEffect, useState } from "react";
import { GetActionPolicies } from "../services/actions";
import { getNodeLabel, Node } from "../services/nodeService";
>>>>>>> main

interface Props {
  node: Node;
}

const ActionOverviewPanel = ({ node }: Props) => {
<<<<<<< HEAD
  return <div>ActionOverviewPanel</div>;
=======
  const [inboundPaths, setInboundPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  useEffect(() => {
    const { request, cancel } = GetActionPolicies(node.properties.map.name);
    request
      .then((res) => {
        setInboundPaths(res.data.map((path) => path));
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching inbound roles:", error);
        }
      });

    return cancel;
  }, [node.properties.map.roleid]);

  return (
    <Box>
      <Accordion allowMultiple={true} width="100%">
        <PathAccordionList
          paths={inboundPaths}
          name="Policies with Action"
          pathFunction={(n) => {
            addPathToGraph(n, addNode, addEdge);
          }}
          pathLabelFunction={(n) => getNodeLabel(n.Nodes[n.Nodes.length - 1])}
        ></PathAccordionList>
      </Accordion>
    </Box>
  );
>>>>>>> main
};

export default ActionOverviewPanel;
