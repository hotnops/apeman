import { useEffect, useState } from "react";
import nodeService, { Node, kinds } from "../services/nodeService";
import { Accordion } from "@chakra-ui/react";
import AccordionList from "./AccordionList";
import DirectTierZeroList from "./DirectTierZeroList";
import IndirectTierZeroList from "./IndirectTierZeroList";

interface Props {
  node: Node;
}

function getNodeWithParams<T>(
  params: Map<string, string>,
  setNodes: (nodes: T[]) => void
) {
  const { request, cancel } = nodeService.getNodesWithParams<T>(params);

  request
    .then((res) => {
      setNodes(res.data);
    })
    .catch(() => {
      cancel();
    });
}

const AccountOverviewPanel = ({ node }: Props) => {
  const [users, setUsers] = useState<Node[]>([]);
  const [groups, setGroups] = useState<Node[]>([]);
  const [policies, setPolicies] = useState<Node[]>([]);
  const [roles, setRoles] = useState<Node[]>([]);

  useEffect(() => {
    if (node) {
      const account_id = node.properties.map.account_id;
      getNodeWithParams<Node>(
        new Map<string, string>([
          ["kind", kinds.AWSRole],
          ["account_id", account_id],
        ]),
        setRoles
      );
      getNodeWithParams<Node>(
        new Map<string, string>([
          ["kind", kinds.AWSManagedPolicy],
          ["account_id", account_id],
        ]),
        setPolicies
      );
      getNodeWithParams<Node>(
        new Map<string, string>([
          ["kind", kinds.AWSUser],
          ["account_id", account_id],
        ]),
        setUsers
      );
      getNodeWithParams<Node>(
        new Map<string, string>([
          ["kind", kinds.AWSGroup],
          ["account_id", account_id],
        ]),
        setGroups
      );
    }
  }, [node]);

  return (
    <Accordion allowMultiple={true} width="100%">
      <AccordionList name="Roles" nodes={roles} />
      <AccordionList name="Policies" nodes={policies}></AccordionList>
      <AccordionList name="Users" nodes={users}></AccordionList>
      <AccordionList name="Groups" nodes={groups}></AccordionList>
      <DirectTierZeroList
        account_id={node.properties.map.account_id}
      ></DirectTierZeroList>
      <IndirectTierZeroList
        account_id={node.properties.map.account_id}
      ></IndirectTierZeroList>
    </Accordion>
  );
};

export default AccountOverviewPanel;
