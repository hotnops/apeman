import { useEffect, useState } from "react";
import NodeService, { Node } from "../services/nodeService";
import { Accordion, Table, Tbody, Td, Tr } from "@chakra-ui/react";
import RoleService, {
  GetInboundRoles,
  GetOutboundRoles,
} from "../services/roleService";
import AccordionList from "./AccordionList";
import { Path, addPathToGraph } from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import PathAccordionList from "./PathAccordionList";

interface Props {
  node: Node;
}

const RoleOverviewPanel = ({ node }: Props) => {
  const [attachedPolicies, setAttachedPolicies] = useState<Node[]>([]);
  const [inboundPaths, setInboundPaths] = useState<Path[]>([]);
  const [outboundPaths, setOutboundPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  useEffect(() => {
    let isMounted = true; // Flag to prevent setting state if the component is unmounted
    setAttachedPolicies([]);

    const fetchPolicies = async () => {
      try {
        const { request } = RoleService.getRolePolicyNodes(
          node.properties.map.roleid
        );
        const res = await request;

        const policyRequests = res.data.nodes.map((node: Node) => {
          const { request } = NodeService.getNodeByID(node.id.toString());
          return request;
        });

        const responses = await Promise.all(policyRequests);

        if (isMounted) {
          const newPolicies = responses.map((res) => res.data as Node);
          setAttachedPolicies(newPolicies);
        }
      } catch (error) {
        console.error("Error fetching policies:", error);
      }
    };

    fetchPolicies();

    return () => {
      isMounted = false; // Prevent state updates if the component is unmounted
      // Add any necessary cleanup here
    };
  }, []); // Add node.properties.map.roleid as a dependency if it can change

  useEffect(() => {
    const { request, cancel } = GetInboundRoles(node.properties.map.roleid);
    request.then((res) => {
      setInboundPaths(res.data.map((path: Path) => path));
    });
    return cancel;
  }, []);

  useEffect(() => {
    const { request, cancel } = GetOutboundRoles(node.properties.map.roleid);
    request.then((res) => {
      setOutboundPaths(res.data.map((path: Path) => path));
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
        <PathAccordionList
          paths={inboundPaths}
          name="Inbound Principals"
          pathFunction={(n: Path) => {
            addPathToGraph(n, addNode, addEdge);
          }}
          pathLabelFunction={(n: Path) => n.Nodes[0].properties.map.arn}
        ></PathAccordionList>
        <PathAccordionList
          paths={outboundPaths}
          name="Outbound Principals"
          pathFunction={(n: Path) => {
            addPathToGraph(n, addNode, addEdge);
          }}
          pathLabelFunction={(n: Path) =>
            n.Nodes?.[1]?.properties?.map?.arn || "Unknown"
          }
        ></PathAccordionList>
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
