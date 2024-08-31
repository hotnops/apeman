import {
  Box,
  Tabs,
  TabList,
  Tab,
  TabPanel,
  TabPanels,
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Table,
  TableContainer,
  Tbody,
  Td,
  Text,
  Tooltip,
  Tr,
  IconButton,
  Skeleton,
  HStack,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { PiGraph } from "react-icons/pi";
import nodeService, {
  Node,
  getIconURL,
  getNodeLabel,
} from "../services/nodeService";
import resourceService from "../services/resourceService";
import {
  GetNodePermissionPathWithAction,
  Path,
  addPathToGraph,
} from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { GetInboundPaths } from "../services/roleService";
import React from "react";

interface Props {
  node: Node;
}

interface PermissionPath {
  node: Node;
  via: Node[];
}

const ResourceOverview = ({ node }: Props) => {
  const [actions, setActions] = useState<string[]>([]);
  const [principals, setPrincipals] = useState<Node[]>([]);
  const [actionPanelData, setActionPanelData] = useState<{
    [action: string]: PermissionPath[];
  }>({});
  const [principalPanelData, setPrincipalPanelData] = useState<{
    [principalArn: string]: string[];
  }>({});
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
  const { addNode, addEdge } = useApemanGraph();

  const handlePrincipalToResourceWithActionPathClick = (
    principalId: number,
    resourceId: number,
    action: string
  ) => {
    const { request, cancel } = GetNodePermissionPathWithAction(
      principalId,
      resourceId,
      action
    );

    request
      .then((res) => {
        res.data.forEach((path) => {
          addPathToGraph(path, addNode, addEdge);
        });
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching node permissions:", error);
        }
      });

    return cancel;
  };

  const handleActionAccordionChange = async (index: number | number[]) => {
    const newIndex = Array.isArray(index) ? index[0] : index;

    index = Array.isArray(index) ? index : [index];

    index.forEach((i) => {
      const action = actions[i];

      if (!actionPanelData[action]) {
        const { request, cancel } =
          resourceService.getPrincipalsWithResourceAndAction(node, action);

        request
          .then((res) => {
            setActionPanelData((prev) => ({
              ...prev,
              [action]: res.data.map((prinNode: Node) => ({
                node: prinNode,
                via: [],
              })),
            }));
            res.data.map((prinNode: Node) => {
              if (prinNode.kinds.includes("AWSRole")) {
                const roleId = prinNode.properties.map["roleid"];
                const { request, cancel } = GetInboundPaths(roleId);

                request
                  .then((res) => {
                    res.data.forEach((path: Path) => {
                      console.log("Inbound path: ", path);
                      const pathNodes = path.Nodes.slice(1); // Get all nodes except the last
                      const lastNode = path.Nodes[0]; // Get the last node

                      setActionPanelData((prev) => ({
                        ...prev,
                        [action]: [
                          ...(prev[action] || []), // Preserve existing data for this action
                          { node: lastNode, via: pathNodes },
                        ],
                      }));
                    });
                  })
                  .catch((error) => {
                    console.error("Error fetching paths:", error);
                  });

                return cancel;
              }
            });
          })
          .catch((error) => {
            if (error.code !== "ERR_CANCELED") {
              console.error(
                "Error fetching resource action principals:",
                error
              );
            }
          });

        return cancel;
      }
    });

    if (newIndex !== expandedIndex) {
      setExpandedIndex(newIndex);
    }
  };

  const handlePrincipalAccordionChange = async (index: number | number[]) => {
    const newIndex = Array.isArray(index) ? index[0] : index;
    index = Array.isArray(index) ? index : [index];

    index.forEach((i) => {
      const principalArn = principals[i].properties.map["arn"];

      if (!principalPanelData[principalArn]) {
        const { request, cancel } =
          resourceService.getActionsWithResourceAndPrincipal(
            node,
            principalArn
          );

        request
          .then((res) => {
            setPrincipalPanelData((prev) => ({
              ...prev,
              [principalArn]: res.data,
            }));
          })
          .catch((error) => {
            if (error.code !== "ERR_CANCELED") {
              console.error(
                "Error fetching resource action principals:",
                error
              );
            }
          });

        return cancel;
      }
    });

    if (newIndex !== expandedIndex) {
      setExpandedIndex(newIndex);
    }
  };

  useEffect(() => {
    let isMounted = true;

    const {
      request: getResourceActionsRequest,
      cancel: cancelGetResourceActions,
    } = resourceService.getResourceActions(node);
    const {
      request: getInboundPermissionsRequest,
      cancel: cancelGetInboundPermissions,
    } = resourceService.getInboundResourcePermissions(node);

    Promise.all([getResourceActionsRequest, getInboundPermissionsRequest])
      .then(([actionsRes, permissionsRes]) => {
        if (isMounted) {
          setActions(
            actionsRes.data.sort((a: string, b: string) => a.localeCompare(b))
          );

          permissionsRes.data.map((principalArn) => {
            nodeService.getNodeByArn(principalArn).request.then((res) => {
              if (isMounted) {
                setPrincipals((prev) => [...prev, res.data[0]]);
              }
            });
          });
        }
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error(
            "Error fetching resource actions or permissions:",
            error
          );
        }
      });

    return () => {
      isMounted = false;
      cancelGetResourceActions();
      cancelGetInboundPermissions();
    };
  }, [node]);

  useEffect(() => {
    console.log("actionPanelData updated:", actionPanelData);
  }, [actionPanelData]);

  return (
    <Box height="100vh" display="flex" flexDirection="column">
      <HStack justifyContent={"space-between"}>
        <Text fontSize="md" as="b">
          Inbound Permissions
        </Text>
      </HStack>
      <Tabs flex="1" display="flex" flexDirection="column">
        <TabList>
          <Tab>Actions</Tab>
          <Tab>Principals</Tab>
        </TabList>
        <TabPanels flex="1" display="flex" flexDirection="column">
          <TabPanel flex="1">
            <Accordion allowMultiple onChange={handleActionAccordionChange}>
              {actions.map((action) => (
                <AccordionItem key={action}>
                  <h2>
                    <AccordionButton>
                      <Box as="span" flex="1" textAlign="left" width="100%">
                        {action}
                        <AccordionIcon />
                      </Box>
                    </AccordionButton>
                  </h2>
                  <AccordionPanel>
                    {actionPanelData[action] ? (
                      <TableContainer>
                        <Table>
                          <Tbody>
                            {actionPanelData[action].map((permPath) => (
                              <Tr key={permPath.node.id}>
                                <Td width="10px">
                                  <Box boxSize="10px">
                                    <img
                                      src={getIconURL(permPath.node.kinds)}
                                      alt="icon"
                                      width="10px"
                                    />
                                  </Box>
                                </Td>
                                <Td>
                                  <Tooltip
                                    label={getNodeLabel(permPath.node)}
                                    hasArrow
                                  >
                                    <Text
                                      fontSize="xs"
                                      maxWidth="200px"
                                      whiteSpace="nowrap"
                                      //overflow="hidden"
                                      //textOverflow="ellipsis"
                                    >
                                      {getNodeLabel(permPath.node)}
                                      {permPath.via.length > 0 && (
                                        <>
                                          {" "}
                                          <b>via</b>
                                          <br />
                                          {permPath.via.map((node, index) => (
                                            <React.Fragment key={index}>
                                              {getNodeLabel(node)}
                                              {index <
                                                permPath.via.length - 1 && ", "}
                                            </React.Fragment>
                                          ))}
                                        </>
                                      )}
                                    </Text>
                                  </Tooltip>
                                </Td>
                                <Td>
                                  <IconButton
                                    aria-label="Graph permissions"
                                    icon={<PiGraph />}
                                    onClick={() =>
                                      handlePrincipalToResourceWithActionPathClick(
                                        permPath.node.id,
                                        node.id,
                                        action
                                      )
                                    }
                                  />
                                </Td>
                              </Tr>
                            ))}
                          </Tbody>
                        </Table>
                      </TableContainer>
                    ) : (
                      <Skeleton height="100%"></Skeleton>
                    )}
                  </AccordionPanel>
                </AccordionItem>
              ))}
            </Accordion>
          </TabPanel>
          <TabPanel flex="1" display="flex" flexDirection="column">
            <Accordion allowMultiple onChange={handlePrincipalAccordionChange}>
              {principals.map((principalNode) => (
                <AccordionItem key={principalNode.id}>
                  <h2>
                    <AccordionButton>
                      <Box as="span" flex="1" textAlign="left" width="100%">
                        {getNodeLabel(principalNode)}
                        <AccordionIcon />
                      </Box>
                    </AccordionButton>
                  </h2>
                  <AccordionPanel>
                    {principalPanelData[principalNode.properties.map["arn"]] ? (
                      <TableContainer>
                        <Table>
                          <Tbody>
                            {principalPanelData[
                              principalNode.properties.map["arn"]
                            ].map((action) => (
                              <Tr key={action}>
                                <Td>
                                  <Text>{action}</Text>
                                </Td>
                                <Td>
                                  <IconButton
                                    aria-label="Graph permissions"
                                    icon={<PiGraph />}
                                    onClick={() =>
                                      handlePrincipalToResourceWithActionPathClick(
                                        principalNode.id,
                                        node.id,
                                        action
                                      )
                                    }
                                  />
                                </Td>
                              </Tr>
                            ))}
                          </Tbody>
                        </Table>
                      </TableContainer>
                    ) : (
                      <Skeleton height="100%"></Skeleton>
                    )}
                  </AccordionPanel>
                </AccordionItem>
              ))}
            </Accordion>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Box>
  );
};

export default ResourceOverview;
