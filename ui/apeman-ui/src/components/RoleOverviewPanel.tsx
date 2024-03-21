import { useEffect, useState } from "react";
import NodeService, { Node } from "../services/nodeService";
import { Accordion, Table, Tbody, Td, Tr } from "@chakra-ui/react";
import RoleService, { GetInboundRoles } from "../services/roleService";
import AccordionList from "./AccordionList";
import {
  Path,
  addPathToGraph,
  getNodesFromPaths,
} from "../services/pathService";
import HoverIcon from "./HoverIcon";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  node: Node;
}

const RoleOverviewPanel = ({ node }: Props) => {
  const [attachedPolicies, setAttachedPolicies] = useState<Node[]>([]);
  const [inboundPaths, setInboundPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  function graphRolePath(n: Node) {
    // Get Path for node
    console.log(`Node id: ${n.id}`);
    const path = inboundPaths.filter((path) => path.Nodes[0].id == n.id)[0];
    console.log("PATH");
    console.log(path);
    addPathToGraph(path, addNode, addEdge);
  }

  useEffect(() => {
    setAttachedPolicies([]);
    const { request, cancel } = RoleService.getRolePolicyNodes(
      node.properties.map.roleid
    );

    request.then((res) => {
      res.data.nodes.map((node: Node) => {
        const { request } = NodeService.getNodeByID(node.id.toString());
        request.then((res) => {
          setAttachedPolicies((attachedPolicies) => [
            ...attachedPolicies,
            res.data as Node,
          ]);
        });
      });
    });

    return () => {
      cancel();
    };
  }, []);

  useEffect(() => {
    const { request, cancel } = GetInboundRoles(node.properties.map.roleid);
    request.then((res) => {
      setInboundPaths(res.data.map((path: Path) => path));
    });
    return cancel;
  }, []);

  return (
    <>
      <Table>
        <Tbody>
          <Tr key="rolename">
            <Td>Role Name</Td>
            <Td>{node.properties.map.rolename}</Td>
          </Tr>
          <Tr key="roleid">
            <Td>Role ID</Td>
            <Td>{node.properties.map.roleid}</Td>
          </Tr>
          <Tr key="accountid">
            <Td>Account ID</Td>
            <Td>{node.properties.map.account_id}</Td>
          </Tr>
        </Tbody>
      </Table>
      <Accordion allowMultiple={true} width="100%">
        <AccordionList
          nodes={getNodesFromPaths(inboundPaths)}
          name="Inbound Principals"
          pathFunction={(n: Node) => {
            graphRolePath(n);
          }}
        ></AccordionList>
      </Accordion>
      <Accordion allowMultiple={true} width="100%">
        <AccordionList
          nodes={attachedPolicies}
          name="Attached Policies"
        ></AccordionList>
      </Accordion>
    </>
  );
};

export default RoleOverviewPanel;
