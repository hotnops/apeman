import {
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  HStack,
  IconButton,
  List,
  ListItem,
  Text,
} from "@chakra-ui/react";
import {
  GetNodePermissionPathWithAction,
  addPathToGraph,
} from "../services/pathService";
import { PiGraph } from "react-icons/pi";
import nodeService from "../services/nodeService";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  name: string;
  actions: string[];
  resourceId: number;
}

const PermissionItem = ({ name, actions, resourceId }: Props) => {
  const { addEdge, addNode } = useApemanGraph();
  const graphPermissionPathWithAction = (
    name: string,
    resourceId: number,
    action: string
  ) => {
    const { request: getRequest } = nodeService.getNodeByArn(name);

    getRequest
      .then((res) => {
        if (res.data.length == 0) {
          console.error("Node not found");
          return;
        }
        const prinNode = res.data[0];
        const prinId = prinNode.id;
        const { request: getPermissionRequest } =
          GetNodePermissionPathWithAction(prinId, resourceId, action);

        getPermissionRequest
          .then((response) => {
            response.data.map((path) => {
              addPathToGraph(path, addNode, addEdge);
            });
          })
          .catch((error) => {
            console.error("Error in getting node permission path:", error);
          });
      })
      .catch((error) => {
        console.error("Error in getting node by ARN:", error);
      });
  };

  return (
    <AccordionItem key={name}>
      <h2>
        <AccordionButton onClick={populateActions()}>
          <Box as="span" flex="1" textAlign="left">
            <Text fontWeight="bold" fontSize="xs">
              {name}
            </Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
      </h2>
      <AccordionPanel>
        <List spacing={3}>
          {actions.map((action) => (
            <ListItem key={name + action}>
              <HStack width="100%">
                <Text width="90%" fontSize="xs">
                  {action}
                </Text>
                <IconButton
                  aria-label="Graph permissions"
                  onClick={() =>
                    graphPermissionPathWithAction(name, resourceId, action)
                  }
                  icon={<PiGraph />}
                />
              </HStack>
            </ListItem>
          ))}
        </List>
      </AccordionPanel>
    </AccordionItem>
  );
};

export default PermissionItem;
