import React, { useEffect, useState } from "react";

import {
  Box,
  Button,
  Card,
  HStack,
  IconButton,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
} from "@chakra-ui/react";
import { CloseIcon, SmallCloseIcon } from "@chakra-ui/icons";
import { useReactTreeApi } from "@naisutech/react-tree";
import nodeService, {
  Node,
  getIconURL,
  getNodeLabel,
} from "../services/nodeService";
import NodeOverview from "./NodeOverview";
import NodeOverviewPanel from "./NodeOverviewPanel";
import { set } from "lodash";

interface Props {
  graphNodes: { [key: string]: Node };
  setGraphNodes: (nodes: { [key: string]: Node }) => void;
}

const NodeBar = ({ graphNodes, setGraphNodes }: Props) => {
  const treeApi = useReactTreeApi();

  const closeTab = (nodeId: string) => {
    setGraphNodes((graphNodes) => {
      const newNodes = { ...graphNodes };
      delete newNodes[nodeId];
      console.log("Setting new nodes");
      console.log(newNodes);
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
            <Tab key={nodeId}>
              <HStack>
                <Box boxSize="10px">
                  <img src={getIconURL(graphNodes[nodeId].kinds)} />
                </Box>
                <Text>{getNodeLabel(graphNodes[nodeId])}</Text>
                <IconButton
                  aria-label="Close tab"
                  icon={<SmallCloseIcon />}
                  size={"xs"}
                  onClick={() => closeTab(nodeId)}
                />
              </HStack>
            </Tab>
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
