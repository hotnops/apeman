import {
  Card,
  HStack,
  useTheme,
  Tabs,
  TabList,
  Tab,
  TabPanel,
  TabPanels,
  IconButton,
} from "@chakra-ui/react";
import { Node } from "../services/nodeService";
import { useEffect, useRef, useState } from "react";
import { useApemanGraph } from "../hooks/useApemanGraph";
import {
  GetNodePermissionPath,
  addPathToGraph,
  Path,
  GetNodeShortestPath,
  GetNodeIdentityPath,
} from "../services/pathService";
import { FaUser } from "react-icons/fa";
import { MdChecklist } from "react-icons/md";
import { FaArrowsAltH } from "react-icons/fa";
import PermissionPathFinder from "./PermissionPathFinder";
import { IoClose } from "react-icons/io5";

interface Props {
  onClose: () => void;
}

const Pathfinder = ({ onClose }: Props) => {
  const [pathNodes, setPathNodes] = useState<Node[]>([]);
  const [tabIndex, setTabIndex] = useState(0);
  const wrapperRef = useRef(null);
  const { addNode, addEdge } = useApemanGraph();
  const theme = useTheme();

  useEffect(() => {
    if (pathNodes.length < 2) {
      return;
    }

    if (tabIndex == 0) {
      const { request, cancel } = GetNodeIdentityPath(
        pathNodes[0].id,
        pathNodes[1].id
      );
      request.then((response) => {
        response.data.map((path: Path) =>
          addPathToGraph(path, addNode, addEdge)
        );
      });

      return cancel;
    } else if (tabIndex == 1) {
      const { request, cancel } = GetNodePermissionPath(
        pathNodes[0].id,
        pathNodes[1].id
      );
      request.then((response) => {
        response.data.map((path: Path) =>
          addPathToGraph(path, addNode, addEdge)
        );
      });

      return cancel;
    } else {
      const { request, cancel } = GetNodeShortestPath(
        pathNodes[0].id,
        pathNodes[1].id
      );
      request.then((response) => {
        response.data.map((path: Path) =>
          addPathToGraph(path, addNode, addEdge)
        );
      });

      return cancel;
    }
  }, [pathNodes]);

  return (
    <div ref={wrapperRef}>
      <Card
        position="relative"
        top="15px"
        left="10px"
        width="30em"
        zIndex={1}
        margin="5px"
        backgroundColor={theme.colors.white}
      >
        <HStack justifyContent="right">
          <IconButton
            aria-label="close nav"
            icon={<IoClose></IoClose>}
            isRound={true}
            onClick={onClose}
            size="xs"
          />
        </HStack>
        <Tabs onChange={(index) => setTabIndex(index)}>
          <TabList>
            <Tab>
              <FaUser></FaUser>
            </Tab>
            <Tab>
              <MdChecklist />
            </Tab>
            <Tab>
              <FaArrowsAltH />
            </Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <PermissionPathFinder
                nodes={pathNodes}
                setPathNodes={setPathNodes}
              ></PermissionPathFinder>
            </TabPanel>
            <TabPanel>
              <PermissionPathFinder
                nodes={pathNodes}
                setPathNodes={setPathNodes}
              ></PermissionPathFinder>
            </TabPanel>
            <TabPanel>
              <PermissionPathFinder
                nodes={pathNodes}
                setPathNodes={setPathNodes}
              ></PermissionPathFinder>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Card>
    </div>
  );
};

export default Pathfinder;
