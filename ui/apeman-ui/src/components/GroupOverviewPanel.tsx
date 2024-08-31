import { useEffect, useState } from "react";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { Path, addPathToGraph } from "../services/pathService";
import { Node } from "../services/nodeService";
import { Accordion, Card, Table, Tbody, Td, Text, Tr } from "@chakra-ui/react";
import PathAccordionList from "./PathAccordionList";
//import AccordionList from "./AccordionList";

import InlinePolicy from "./InlinePolicy";
import groupService from "../services/groupService";
import policyService from "../services/policyService";

interface Props {
  node: Node;
}

const GroupOverviewPanel = ({ node }: Props) => {
  const [membershipPaths, setMembershipPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();
  const [inlinePolicy, setInlinePolicy] = useState<Node | null>(null);

  useEffect(() => {
    const resp = groupService.getGroupMembershipPaths(
      node.properties.map.groupid
    );
    resp.request.then((res) => {
      setMembershipPaths(res.data.map((path: Path) => path));
    });
  }, [node]);

  useEffect(() => {
    const { request, cancel } = policyService.getInlinePolicyNode(node);

    request?.then((res) => {
      setInlinePolicy(res.data);
    });

    return cancel;
  }, [node]);

  return (
    <>
      <Card>
        <Table size={"xs"} variant="unstytled">
          <Tbody>
            <Tr key="groupname">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  Group Name
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.name}
                </Text>
              </Td>
            </Tr>
            <Tr key="userid">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  Group ID
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.groupid}
                </Text>
              </Td>
            </Tr>
            <Tr key="accountid">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  Account ID
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.account_id}
                </Text>
              </Td>
            </Tr>
          </Tbody>
        </Table>
      </Card>
      <Card marginY="25px">
        <Accordion allowMultiple={true} width="100%">
          <PathAccordionList
            paths={membershipPaths}
            name="Group Memberships"
            pathFunction={(n) => {
              addPathToGraph(n, addNode, addEdge);
            }}
            pathLabelFunction={(n) => n.Nodes[0].properties.map.arn}
          ></PathAccordionList>
        </Accordion>
        {/* <Accordion allowMultiple={true} width="100%">
          <AccordionList
            nodes={attachedPolicies}
            name="Attached Policies"
          ></AccordionList>
        </Accordion> */}
        {inlinePolicy && <InlinePolicy node={inlinePolicy} />}
      </Card>
    </>
  );
};

export default GroupOverviewPanel;
