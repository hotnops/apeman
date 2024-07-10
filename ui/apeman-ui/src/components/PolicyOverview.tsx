import { Accordion, Code } from "@chakra-ui/react";
import PolicyService from "../services/policyService";
import { useEffect, useState } from "react";
import { Node } from "../services/nodeService";
import AccordionList from "./AccordionList";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";

interface Props {
  node: Node;
}

const PolicyOverview = ({ node }: Props) => {
  const [attachedPrincipals, setAttachedPrincipals] = useState<Node[]>([]);
  const [policyObject, setPolicyObject] = useState(null);
  useEffect(() => {
    // Get all the principals that are attached to the policy
    const { request, cancel } = PolicyService.getPolicyPrincipalNodes(
      node.properties.map.policyid
    );
    request.then((res) => {
      console.log(typeof res.data);
      const newPrincipals = res.data.map((prinNode: Node) => prinNode);

      // Update state once with the new principals
      setAttachedPrincipals((attachedPrincipals) => [
        ...attachedPrincipals,
        ...newPrincipals,
      ]);
    });
  }, []);

  useEffect(() => {
    // Get the policy object
    const { request, cancel } = PolicyService.getManagedPolicyJSON(
      node.properties.map.policyid
    );
    request.then((res) => {
      console.log(typeof res.data);
      setPolicyObject(res.data);
    });

    return () => {
      cancel();
    };
  }, []);

  return (
    <Accordion allowMultiple={true}>
      <AccordionList
        nodes={attachedPrincipals}
        name="Attached Principals"
      ></AccordionList>
      <SyntaxHighlighter language="json" style={coy}>
        {policyObject && JSON.stringify(policyObject, null, 4)}
      </SyntaxHighlighter>
    </Accordion>
  );
};

export default PolicyOverview;
