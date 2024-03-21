import {
  Box,
  Card,
  HStack,
  Input,
  useTheme,
  Text,
  IconButton,
  Tabs,
  TabList,
  Tab,
  TabPanel,
  TabPanels,
} from "@chakra-ui/react";
import NodeSuggestions from "./NodeSuggestions";
import { Node, getNodeLabel } from "../services/nodeService";
import { useEffect, useRef, useState } from "react";
import SearchBar from "./SearchBar";
import { MdTripOrigin } from "react-icons/md";
import { MdOutlinePinDrop } from "react-icons/md";
import { IoClose } from "react-icons/io5";
import HoverIcon from "./HoverIcon";
import { RiDirectionLine } from "react-icons/ri";
import { IoMdArrowRoundDown } from "react-icons/io";
import { IoEllipsisVerticalSharp, IoCloseCircleOutline } from "react-icons/io5";
import NodeListItem from "./NodeListItem";
import { useApemanGraph } from "../hooks/useApemanGraph";
import {
  GetNodePermissionPath,
  addPathToGraph,
  Path,
  GetNodeShortestPath,
  GetNodeIdentityPath,
} from "../services/pathService";
import { Relationship } from "../services/relationshipServices";
import { Icon } from "reagraph";
import { FaUser } from "react-icons/fa";
import { MdChecklist } from "react-icons/md";
import { FaArrowsAltH } from "react-icons/fa";
import PermissionPathFinder from "./PermissionPathFinder";

interface Props {
  onClose: () => void;
}

const Pathfinder = ({ onClose }: Props) => {
  const [pathNodes, setPathNodes] = useState<Node[]>([]);
  const [tabIndex, setTabIndex] = useState(0);
  const wrapperRef = useRef(null);
  const { addNode, addEdge } = useApemanGraph();
  const theme = useTheme();

  //   useEffect(() => {
  //     document.addEventListener("mousedown", handleFocusChange);
  //     return () => {
  //       document.removeEventListener("mousedown", handleFocusChange);
  //     };
  //   });

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
        console.log(response.data);
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
        console.log(response.data);
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
        console.log(response.data);
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
        position="fixed"
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
