import React, { useState } from "react";
import styled from "styled-components";
import ApemanGraph from "./components/ApemanGraph";
import NodeExplorer from "./components/NodeExplorer";

import NodeBar from "./components/NodeBar";
import { Node } from "./services/nodeService";
import { Tab, TabList, Tabs } from "@chakra-ui/react";
import NavBar from "./components/NavBar";
import { useApemanGraph } from "./hooks/useApemanGraph";
import Pathfinder from "./components/Pathfinder";

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
  right: 0;
  top: 0;
  bottom: 0;
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

  const [graphNodes, setGraphNodes] = useState<{ [key: string]: Node }>({});
  const [showPathfinder, setShowPathfinder] = useState(false);
  const { activeElement } = useApemanGraph();

  const handleMouseDown = (e: React.MouseEvent) => {
    const startX = e.clientX;
    const startWidth = panelWidth;

    const onMouseMove = (e: MouseEvent) => {
      const newWidth = startWidth + (e.clientX - startX);
      setPanelWidth(newWidth > 100 ? newWidth : 100); // Minimum width of 100px
    };

    const onMouseUp = () => {
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  };

  return (
    <Container>
      <Header>
        <Tabs variant={"soft-rounded"} padding="10px">
          <TabList>
            <Tab>Explorer</Tab>
            <Tab>Context Manager</Tab>
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
            {!activeElement &&
              (showPathfinder ? (
                <Pathfinder
                  onClose={() => setShowPathfinder(false)}
                ></Pathfinder>
              ) : (
                <NavBar closeNavBar={() => setShowPathfinder(true)}></NavBar>
              ))}
            <ApemanGraph setGraphNodes={setGraphNodes} />
          </ApemanGraphContainer>
        </MainContent>
      </MainContainer>
    </Container>
  );
};

export default App;
