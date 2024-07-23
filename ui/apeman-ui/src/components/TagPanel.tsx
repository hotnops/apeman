import {
  Table,
  TableContainer,
  Td,
  Tr,
  Text,
  Thead,
  Th,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import nodeService, { Node } from "../services/nodeService";
import { useApemanGraph } from "../hooks/useApemanGraph";

const TagPanel = () => {
  const [tagNodes, setTagNodes] = useState<Node[]>([]);

  const { activeElement: activeNode } = useApemanGraph();
  useEffect(() => {
    if (activeNode == null) {
      return;
    }
    const { request, cancel } = nodeService.getNodeTags(
      (activeNode as Node).id.toString()
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
            <Thead>
              <Th>
                <Text as="b">Key</Text>
              </Th>
              <Th>
                <Text as="b">Value</Text>
              </Th>
            </Thead>
            {tagNodes.map((tagNode) => (
              <Tr>
                <Td key={tagNode.properties.map.key}>
                  <Text>{tagNode.properties.map.key}</Text>
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
