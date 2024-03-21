import {
  Box,
  Button,
  Card,
  Drawer,
  DrawerBody,
  DrawerContent,
  DrawerHeader,
  DrawerOverlay,
  HStack,
  Spinner,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import NavBar from "./components/NavBar";
import { useEffect, useState } from "react";
import ApemanGraph from "./components/ApemanGraph";
import nodeService, { Node, nodeToGraphNode } from "./services/nodeService";
import NodeOverviewPanel from "./components/NodeOverviewPanel";
import { kinds } from "./services/nodeService";
import { useApemanGraph } from "./hooks/useApemanGraph";
import Pathfinder from "./components/Pathfinder";
import { IoTrashOutline } from "react-icons/io5";
import HoverIcon from "./components/HoverIcon";
import { useTheme } from "@emotion/react";
import { SettingsIcon } from "@chakra-ui/icons";
import ContextSettings from "./components/ContextSettings";
import EdgeOverviewPanel from "./components/EdgeOverviewPanel";

function App() {
  // First time render, simply show the AWS accounts[
  const { addNode, activeElement, setNodes, setActiveElement, setEdges } =
    useApemanGraph();
  const [showPathfinder, setShowPathfinder] = useState(false);
  const theme = useTheme();

  useEffect(() => {
    const { request, cancel } = nodeService.getNodesByKind(kinds.AWSAccount);

    request.then((res) => {
      res.data.map((node) => {
        addNode(node as Node);
      });
    });

    return () => {
      cancel();
    };
  }, []);

  const { isOpen, onClose, onOpen } = useDisclosure();

  return (
    <>
      {!activeElement &&
        (showPathfinder ? (
          <Pathfinder onClose={() => setShowPathfinder(false)}></Pathfinder>
        ) : (
          <NavBar closeNavBar={() => setShowPathfinder(true)}></NavBar>
        ))}

      <Card height="100vh" position="relative" zIndex={0}>
        <ApemanGraph />
      </Card>
      <HStack position="fixed" top={5} right={5} zIndex={1}>
        <Box height="2em" width="2em">
          <HoverIcon
            iconColor={theme.colors.gray[500]}
            hoverColor={theme.colors.gray[900]}
          >
            <IoTrashOutline
              size="100%"
              onClick={() => {
                setActiveElement(null);
                setNodes([]);
                setEdges([]);
              }}
            ></IoTrashOutline>
          </HoverIcon>
        </Box>
        <HoverIcon
          iconColor={theme.colors.gray[500]}
          hoverColor={theme.colors.gray[900]}
        >
          <SettingsIcon boxSize="2em" onClick={onOpen}></SettingsIcon>
        </HoverIcon>
      </HStack>
      {activeElement ? (
        <Card
          position="fixed"
          height="60vh"
          top="60px"
          left={0}
          overflow="scroll"
          zIndex={1}
          minWidth="35vw"
          maxWidth="50vw"
        >
          {"id" in activeElement && (
            <NodeOverviewPanel node={activeElement}></NodeOverviewPanel>
          )}
          {"ID" in activeElement && (
            <EdgeOverviewPanel edge={activeElement}></EdgeOverviewPanel>
          )}
        </Card>
      ) : null}
      <Drawer placement="right" onClose={onClose} isOpen={isOpen} size="lg">
        <DrawerOverlay />
        <DrawerContent width="40vw">
          <DrawerHeader>Context Manager</DrawerHeader>
          <DrawerBody>
            <ContextSettings></ContextSettings>
          </DrawerBody>
        </DrawerContent>
      </Drawer>
    </>
  );
}

export default App;
