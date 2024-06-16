import { Accordion } from "@chakra-ui/react";
import PolicyService from "../services/policyService";
import { useEffect, useState } from "react";
import { Node } from "../services/nodeService";
import AccordionList from "./AccordionList";

interface Props {
  node: Node;
}

const PolicyOverview = ({ node }: Props) => {
  const [attachedPrincipals, setAttachedPrincipals] = useState<Node[]>([]);
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
    </Accordion>
  );
};

export default PolicyOverview;
