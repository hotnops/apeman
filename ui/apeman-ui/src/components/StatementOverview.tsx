import { Accordion, Table, Tbody, Td, Tr } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { getNodeLabel, Node } from "../services/nodeService";

import AccordionList from "./AccordionList";
import statementService, {
  fetchAllStatementData,
  StatementDetails,
} from "../services/statementService";
import SyntaxHighlighter from "react-syntax-highlighter/dist/esm/default-highlight";
import PathAccordionList from "./PathAccordionList";
import { addPathToGraph, Path } from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  node: Node;
}

const StatementOverview = ({ node }: Props) => {
  const [statementObject, setStatementObject] = useState<string>();
  const { addNode, addEdge } = useApemanGraph();
  const [policyPaths, setPolicyPaths] = useState<Path[]>([]);

  useEffect(() => {
    const { request: objectRequest, cancel: objectCancel } =
      statementService.getStatementObject(node.properties.map["hash"]);

    const { request: policyRequest, cancel: policyCancel } =
      statementService.getAttachedPolicies(node.properties.map["hash"]);

    Promise.all([objectRequest, policyRequest]).then(
      ([objectRes, policyRes]) => {
        setStatementObject(objectRes.data);

        setPolicyPaths(policyRes.data);
      }
    );

    return () => {
      objectCancel();
      policyCancel();
    };
  }, [node]);

  // List of policies that use the statement
  // List of actions
  // List of resources
  // List of conditions

  return (
    <>
      <Accordion allowMultiple={true} width="100%">
        <PathAccordionList
          paths={policyPaths}
          name="Attached Policies"
          pathFunction={(n) => {
            addPathToGraph(n, addNode, addEdge);
          }}
          pathLabelFunction={(n) => getNodeLabel(n.Nodes[n.Nodes.length - 1])}
        ></PathAccordionList>
      </Accordion>
      <SyntaxHighlighter>
        {statementObject ? JSON.stringify(statementObject, null, 4) : ""}
      </SyntaxHighlighter>
    </>
  );
};

export default StatementOverview;
