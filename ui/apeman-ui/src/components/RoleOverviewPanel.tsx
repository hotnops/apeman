import { useEffect, useState } from "react";
import NodeService, { Node } from "../services/nodeService";
import { Accordion, Table, Tbody, Text, Td, Tr, Card } from "@chakra-ui/react";
import RoleService, {
  GetInboundRoles,
  GetOutboundRoles,
} from "../services/roleService";
import AccordionList from "./AccordionList";
import { Path, addPathToGraph } from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import PathAccordionList from "./PathAccordionList";
import InlinePolicy from "./InlinePolicy";
import RSOPPanel from "./RSOPPanel";
import AssumeRolePolicyPanel from "./AssumeRolePolicyPanel";
import policyService from "../services/policyService";

interface Props {
  node: Node;
}

const RoleOverviewPanel = ({ node }: Props) => {
  const [attachedPolicies, setAttachedPolicies] = useState<Node[]>([]);
  const [inboundPaths, setInboundPaths] = useState<Path[]>([]);
  const [outboundPaths, setOutboundPaths] = useState<Path[]>([]);
  const [inlinePolicy, setInlinePolicy] = useState<Node | null>(null);
  const { addNode, addEdge } = useApemanGraph();

  useEffect(() => {
    let isMounted = true;
    setAttachedPolicies([]);

    const fetchManagedPolicies = async () => {
      try {
        const { request } = RoleService.getRoleManagedPolicyNodes(
          node.properties.map.roleid
        );
        const res = await request;

        const policyRequests = res.data.nodes.map((node) => {
          const { request } = NodeService.getNodeByID(node.id.toString());
          return request;
        });

        const responses = await Promise.all(policyRequests);

        if (isMounted) {
          const newPolicies = responses.map((res) => res.data);
          setAttachedPolicies(newPolicies);
        }
      } catch (error) {
        console.error("Error fetching policies:", error);
      }
    };

    fetchManagedPolicies();

    return () => {
      isMounted = false;
    };
  }, [node]);

  useEffect(() => {
    const { request, cancel } = GetInboundRoles(node.properties.map.roleid);
    request
      .then((res) => {
        setInboundPaths(res.data.map((path) => path));
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching inbound roles:", error);
        }
      });

    return cancel;
  }, [node.properties.map.roleid]);

  useEffect(() => {
    const { request, cancel } = GetOutboundRoles(node.properties.map.roleid);
    request
      .then((res) => {
        setOutboundPaths(res.data.map((path) => path));
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching outbound roles:", error);
        }
      });

    return cancel;
  }, [node.properties.map.roleid]);

  useEffect(() => {
    console.log("Role node: " + node);
    const { request, cancel } = policyService.getInlinePolicyNode(node);

    request?.then((res) => {
      setInlinePolicy(res.data);
    });

    return cancel;
  }, [node]);

  return (
    <>
      <Card>
        <Table size="xs" variant="unstyled">
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
            pathFunction={(n) => {
              addPathToGraph(n, addNode, addEdge);
            }}
            pathLabelFunction={(n) => n.Nodes[0].properties.map.arn}
          ></PathAccordionList>
          <PathAccordionList
            paths={outboundPaths}
            name="Outbound Principals"
            pathFunction={(n) => {
              addPathToGraph(n, addNode, addEdge);
            }}
            pathLabelFunction={(n) =>
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
        {inlinePolicy && <InlinePolicy node={inlinePolicy}></InlinePolicy>}
        <AssumeRolePolicyPanel roleNode={node} />
      </Card>
      <Card>
        <RSOPPanel node={node} />
      </Card>
    </>
  );
};

export default RoleOverviewPanel;
