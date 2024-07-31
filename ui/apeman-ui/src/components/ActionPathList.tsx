import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Button,
  HStack,
  IconButton,
  Table,
  Tbody,
  Td,
  Text,
  Tr,
} from "@chakra-ui/react";
import React from "react";

import {
  ActionPathEntry,
  GetNodePermissionPathWithAction,
  addPathToGraph,
} from "../services/pathService";
import { PiGraph } from "react-icons/pi";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  actionPathMap: { [key: string]: ActionPathEntry[] };
}

const ActionPathList = ({ actionPathMap }: Props) => {
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

  return (
    <Accordion allowMultiple={true} width="100%">
      {Object.keys(actionPathMap).map((action) => (
        <AccordionItem>
          <h2>
            <AccordionButton>
              <Box flex="1" textAlign="left">
                {action}
              </Box>
              <AccordionIcon />
            </AccordionButton>
          </h2>
          <AccordionPanel>
            <Table size={"xs"}>
              <Tbody>
                {actionPathMap[action].map((entry) => (
                  <Tr key={entry.principal_id}>
                    <Td>
                      <Text fontSize="xs" textOverflow="ellipsis">
                        {entry.resource_arn}
                      </Text>
                    </Td>
                    <Td>
                      <IconButton
                        aria-label="Graph permissions"
                        icon={<PiGraph />}
                        onClick={() =>
                          handlePrincipalToResourceWithActionPathClick(
                            entry.principal_id,
                            entry.resource_id,
                            action
                          )
                        }
                      />
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </AccordionPanel>
        </AccordionItem>
      ))}
    </Accordion>
  );
};

export default ActionPathList;
