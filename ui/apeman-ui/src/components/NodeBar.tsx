import {
  Box,
  HStack,
  IconButton,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
} from "@chakra-ui/react";
import { SmallCloseIcon } from "@chakra-ui/icons";
import { Node, getIconURL, getNodeLabel } from "../services/nodeService";
import NodeOverviewPanel from "./NodeOverviewPanel";

interface Props {
  graphNodes: { [key: string]: Node };
  setGraphNodes: React.Dispatch<React.SetStateAction<{ [key: string]: Node }>>;
}

const NodeBar = ({ graphNodes, setGraphNodes }: Props) => {
  const closeTab = (nodeId: string) => {
    setGraphNodes((graphNodes) => {
      const newNodes = { ...graphNodes };
      delete newNodes[nodeId];
      return newNodes;
    });
  };

  if (graphNodes == null) {
    return null;
  }

  return (
    <Box overflowY="scroll" width="40vw">
      <Tabs size="sm" variant="enclosed">
        <TabList overflowY="auto">
          {Object.keys(graphNodes).map((nodeId) => (
            <Box key={nodeId} position="relative">
              <Tab>
                <HStack>
                  <Box boxSize="10px">
                    <img
                      src={getIconURL(graphNodes[nodeId].kinds)}
                      alt="icon"
                    />
                  </Box>
                  <Text>{getNodeLabel(graphNodes[nodeId])}</Text>
                </HStack>
              </Tab>
              <IconButton
                aria-label="Close tab"
                icon={<SmallCloseIcon />}
                size="xs"
                position="absolute"
                top="2px"
                right="2px"
                onClick={() => closeTab(nodeId)}
              />
            </Box>
          ))}
        </TabList>
        <TabPanels>
          {Object.keys(graphNodes).map((nodeId) => (
            <TabPanel key={nodeId}>
              <NodeOverviewPanel node={graphNodes[nodeId]} />
            </TabPanel>
          ))}
        </TabPanels>
      </Tabs>
    </Box>
  );
};

export default NodeBar;
