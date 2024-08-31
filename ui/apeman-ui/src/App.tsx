import React, { useState, useEffect, useCallback } from "react";
import styled from "styled-components";
import ApemanGraph from "./components/ApemanGraph";
import NodeExplorer from "./components/NodeExplorer";
import NodeBar from "./components/NodeBar";
import { Node } from "./services/nodeService";
import { HStack, IconButton, Tab, TabList, Tabs } from "@chakra-ui/react";
import NavBar from "./components/NavBar";
import { useApemanGraph } from "./hooks/useApemanGraph";
import { IoTrashOutline } from "react-icons/io5";

const Container = styled.div`
  display: flex;
  flex-direction: column;
  height: 100vh;
`;

const Header = styled.div`
  width: 100%;
  text-align: center;
  font-size: 1.5rem;
`;

const MainContainer = styled.div`
  display: flex;
  flex-grow: 1;
  overflow: hidden;
`;

const SidePanel = styled.div<{ width: number }>`
  width: ${(props) => props.width}px;
  height: 100%;
  background-color: #f4f4f4;
  position: relative;
  overflow-y: auto;
`;

const Resizer = styled.div`
  width: 5px;
  background-color: #ccc;
  cursor: ew-resize;
  position: absolute;
  right: -2.5px; /* Ensure resizer is above scrollbar */
  top: 0;
  bottom: 0;
  z-index: 10;
`;

const MainContent = styled.div`
  flex-grow: 1;
  background-color: #fff;
  overflow-y: auto;
  display: flex;
`;

const ApemanGraphContainer = styled.div`
  flex-grow: 1;
  max-width: 100%;
  height: 100%;
  overflow: hidden;
  position: relative;
  padding: 10px;
  margin: 10px;
  border-radius: 10px;
`;

const App: React.FC = () => {
  const [panelWidth, setPanelWidth] = useState(300);
  const [isResizing, setIsResizing] = useState(false);
  const [graphNodes, setGraphNodes] = useState<{ [key: string]: Node }>({});
  const { setNodes, setEdges } = useApemanGraph();

  // Handle mouse movement for resizing
  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (isResizing) {
        const newWidth = panelWidth + e.movementX;
        setPanelWidth(newWidth > 100 ? newWidth : 100); // Minimum width of 100px
      }
    },
    [isResizing, panelWidth]
  );

  // Stop resizing on mouse up
  const handleMouseUp = useCallback(() => {
    if (isResizing) {
      setIsResizing(false);
    }
  }, [isResizing]);

  // Add and remove event listeners
  useEffect(() => {
    if (isResizing) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
    } else {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    }

    // Cleanup function
    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isResizing, handleMouseMove, handleMouseUp]);

  // Start resizing on mouse down
  const handleMouseDown = () => {
    setIsResizing(true);
  };

  return (
    <Container>
      <Header>
        <Tabs variant={"soft-rounded"} padding="10px">
          <TabList>
            <Tab>Explorer</Tab>
            {/* <Tab>Context Manager</Tab> */}
          </TabList>
        </Tabs>
      </Header>
      <MainContainer>
        <SidePanel width={panelWidth}>
          <NodeExplorer graphNodes={graphNodes} setGraphNodes={setGraphNodes} />
          <Resizer onMouseDown={handleMouseDown} />
        </SidePanel>
        <MainContent>
          {Object.keys(graphNodes).length > 0 && (
            <NodeBar graphNodes={graphNodes} setGraphNodes={setGraphNodes} />
          )}
          <ApemanGraphContainer>
            <HStack justifyContent="space-between">
              <NavBar></NavBar>
              <IconButton
                aria-label="clear graph nodes"
                icon={<IoTrashOutline />}
                isRound={true}
                onClick={() => {
                  setNodes([]);
                  setEdges([]);
                }}
                size="md"
                position="relative"
                top="15px"
                right="5px"
                zIndex={1}
              />
            </HStack>
            <ApemanGraph setGraphNodes={setGraphNodes} />
          </ApemanGraphContainer>
        </MainContent>
      </MainContainer>
    </Container>
  );
};

export default App;
