import {
  HStack,
  Table,
  TableContainer,
  Td,
  Tr,
  Wrap,
  WrapItem,
  useTheme,
} from "@chakra-ui/react";
import React, { useEffect, useState } from "react";
import { Badge } from "@chakra-ui/react";
import apiClient from "../services/api-client";
import nodeService, { Node } from "../services/nodeService";
import { useApemanGraph } from "../hooks/useApemanGraph";

const getColor = (index: number) => {
  // I love a rainbow...
  const colors = [
    "red",
    "orange",
    "yellow",
    "green",
    "blue",
    "indigo",
    "vioilet",
  ];
  return colors[index % colors.length];
};

const TagPanel = () => {
  const [tagNodes, setTagNodes] = useState<Node[]>([]);

  const { activeElement: activeNode } = useApemanGraph();
  useEffect(() => {
    if (activeNode == null) {
      return;
    }
    const { request, cancel } = nodeService.getNodeTags(
      activeNode.id.toString()
    );
    request.then((res) => {
      const nodes: Node[] = res.data;
      nodes.map((node) => {
        setTagNodes((prev) => [...prev, node]);
      });
    });
    return cancel;
  }, []);
  return (
    <>
      {tagNodes && (
        <TableContainer>
          <Table variant="striped">
            {tagNodes.map((tagNode) => (
              <Tr>
                <Td key={tagNode.properties.map.key}>
                  {tagNode.properties.map.key}
                </Td>
                <Td key={tagNode.properties.map.value}>
                  {tagNode.properties.map.value}
                </Td>
              </Tr>
            ))}
          </Table>
        </TableContainer>
      )}
    </>
  );
};

export default TagPanel;
