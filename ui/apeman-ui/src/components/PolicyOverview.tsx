import { Accordion, Button, HStack, Text } from "@chakra-ui/react";
import PolicyService from "../services/policyService";
import { useEffect, useState } from "react";
import { Node } from "../services/nodeService";
import AccordionList from "./AccordionList";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";
import { PiGraph } from "react-icons/pi";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { Path, addPathToGraph } from "../services/pathService";

interface Props {
  node: Node;
}

const PolicyOverview = ({ node }: Props) => {
  const [attachedPrincipals, setAttachedPrincipals] = useState<Node[]>([]);
  const [policyObject, setPolicyObject] = useState(null);
  const { addNode, addEdge } = useApemanGraph();

  const graphPolicyNodes = (node: Node) => {
    const { request, cancel } = PolicyService.getNodesAttachedToPolicy(
      node.properties.map.policyid,
      "managed"
    );

    request.then((res) => {
      console.log(res.data);
      res.data.map((path: Path) => addPathToGraph(path, addNode, addEdge));
    });
  };

  useEffect(() => {
    // Get all the principals that are attached to the policy
    const { request, cancel } = PolicyService.getPolicyPrincipalNodes(
      node.properties.map.policyid
    );
    request.then((res) => {
      const newPrincipals = res.data.map((prinNode: Node) => prinNode);

      // Update state once with the new principals
      setAttachedPrincipals((attachedPrincipals: Node[]) => [
        ...attachedPrincipals,
        ...newPrincipals,
      ]);
    });

    return () => cancel();
  }, []);

  useEffect(() => {
    // Get the policy object
    const { request, cancel } = PolicyService.getManagedPolicyJSON(
      node.properties.map.policyid
    );
    request.then((res) => {
      setPolicyObject(res.data);
    });

    return () => {
      cancel();
    };
  }, []);

  return (
    <>
      <HStack justifyContent="space-between" paddingY="20px">
        <Text as="b" fontSize={"md"}>
          {node.properties.map["arn"]}
        </Text>
        <Button onClick={() => graphPolicyNodes(node)}>
          <PiGraph />
        </Button>
      </HStack>
      <Accordion allowMultiple={true}>
        <AccordionList
          nodes={attachedPrincipals}
          name="Attached Principals"
        ></AccordionList>
        <SyntaxHighlighter language="json" style={coy}>
          {policyObject ? JSON.stringify(policyObject, null, 4) : ""}
        </SyntaxHighlighter>
      </Accordion>
    </>
  );
};

export default PolicyOverview;
