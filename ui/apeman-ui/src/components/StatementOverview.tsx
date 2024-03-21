import { Accordion, Table, Tbody, Td, Tr } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { Node } from "../services/nodeService";

import AccordionList from "./AccordionList";
import {
  fetchAllStatementData,
  StatementDetails,
} from "../services/statementService";

interface Props {
  node: Node;
}

const StatementOverview = ({ node }: Props) => {
  const [statementDetails, setStatementDetails] = useState<StatementDetails>({
    policies: [],
    actions: [],
    resources: [],
    conditions: [],
  });

  useEffect(() => {
    fetchAllStatementData(node, setStatementDetails);
  }, []);

  // List of policies that use the statement
  // List of actions
  // List of resources
  // List of conditions

  return (
    <>
      <Table>
        <Tbody>
          <Tr key="effect">
            <Td>Effect</Td>
            <Td>{node.properties.map["effect"]}</Td>
          </Tr>
          <Tr></Tr>
        </Tbody>
      </Table>
      <Accordion allowMultiple={true}>
        <AccordionList
          nodes={statementDetails.actions}
          name="actions"
        ></AccordionList>
      </Accordion>
      <Accordion allowMultiple={true}>
        <AccordionList
          nodes={statementDetails.resources}
          name="resources"
        ></AccordionList>
        <Accordion allowMultiple={true}>
          <AccordionList
            nodes={statementDetails.conditions}
            name="conditions"
          ></AccordionList>
        </Accordion>
      </Accordion>
    </>
  );
};

export default StatementOverview;
