import { useEffect, useState } from "react";
import NodeService, { Node } from "../services/nodeService";
import {
  Accordion,
  Table,
  Tbody,
  Text,
  Td,
  Tr,
  Divider,
  Card,
  AccordionItem,
  AccordionButton,
  Box,
  AccordionIcon,
  AccordionPanel,
} from "@chakra-ui/react";
import RoleService, {
  GetInboundRoles,
  GetOutboundRoles,
} from "../services/roleService";
import AccordionList from "./AccordionList";
import { Path, addPathToGraph } from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import PathAccordionList from "./PathAccordionList";
import RSOPPanel from "./RSOPPanel";
import PermissionList from "./PermissionList";
import { AsyncGetInlinePolicyJSON } from "../services/policyService";
import InlinePolicy from "./InlinePolicy";

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

    const fetchManagedPolicies = async () => {
      try {
        const { request } = RoleService.getRoleManagedPolicyNodes(
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

    fetchManagedPolicies();

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
      <Card>
        <Table size={"xs"} variant="unstyled">
          <Tbody>
            <Tr key="rolename">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  Role Name
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.rolename}
                </Text>
              </Td>
            </Tr>
            <Tr key="roleid">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  Role ID
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.roleid}
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
              n.Nodes?.[n.Nodes.length - 1]?.properties?.map?.arn || "Unknown"
            }
          ></PathAccordionList>
        </Accordion>
        <Accordion allowMultiple={true} width="100%">
          <AccordionList
            nodes={attachedPolicies}
            name="Managed Policies"
          ></AccordionList>
        </Accordion>
        <InlinePolicy principalNode={node} />
      </Card>
      <Card>
        <PermissionList
          endpoint={"roles/" + node.properties.map.roleid + "/rsop"}
          resourceId={() => node.id}
        >
          Resultant Set Of Policy
        </PermissionList>
      </Card>
    </>
  );
};

export default RoleOverviewPanel;
