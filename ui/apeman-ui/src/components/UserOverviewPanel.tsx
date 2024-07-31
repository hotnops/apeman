import { useEffect, useState } from "react";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { Path, addPathToGraph } from "../services/pathService";
import nodeService, { Node } from "../services/nodeService";
import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Card,
  Skeleton,
  Table,
  Tbody,
  Td,
  Text,
  Tr,
} from "@chakra-ui/react";
import PathAccordionList from "./PathAccordionList";
import AccordionList from "./AccordionList";
import PermissionList from "./PermissionList";
import UserService, {
  GetOutboundRoles,
  GetUserRSOPActions,
} from "../services/userService";
import InlinePolicy from "./InlinePolicy";
import ActionPathList from "./ActionPathList";
import { RiSpace } from "react-icons/ri";
import RSOPPanel from "./RSOPPanel";

interface Props {
  node: Node;
}

const UserOverviewPanel = ({ node }: Props) => {
  const [attachedPolicies, setAttachedPolicies] = useState<Node[]>([]);
  const [outboundPaths, setOutboundPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  useEffect(() => {
    let isMounted = true; // Flag to prevent setting state if the component is unmounted
    setAttachedPolicies([]);

    const fetchPolicies = async () => {
      try {
        const { request } = UserService.getUserPolicyNodes(
          node.properties.map.userid
        );
        const res = await request;

        const policyRequests = res.data.nodes.map((node: Node) => {
          const { request } = nodeService.getNodeByID(node.id.toString());
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
  }, [node]);

  useEffect(() => {
    const { request, cancel } = GetOutboundRoles(node.properties.map.userid);
    request.then((res) => {
      setOutboundPaths(res.data.map((path: Path) => path));
    });
    return cancel;
  }, []);

  return (
    <>
      <Card>
        <Table size={"xs"} variant="unstytled">
          <Tbody>
            <Tr key="username">
              <Td>
                <Text fontSize="sm" as="b" padding="5px">
                  User Name
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
                  User ID
                </Text>
              </Td>
              <Td>
                <Text fontSize="sm" textAlign="right" padding="5px">
                  {node.properties.map.userid}
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
            name="Attached Policies"
          ></AccordionList>
        </Accordion>
        <InlinePolicy principalNode={node} />
      </Card>
      <Card>
        <RSOPPanel node={node} />
      </Card>
    </>
  );
};

export default UserOverviewPanel;
