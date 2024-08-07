import {
  Table,
  TableContainer,
  Td,
  Tr,
  Text,
  Thead,
  Th,
  Tbody,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import nodeService, { Node } from "../services/nodeService";

interface Props {
  node: Node;
}

const TagPanel = ({ node }: Props) => {
  const [tagNodes, setTagNodes] = useState<Node[]>([]);

  useEffect(() => {
    let isMounted = true;

    const { request, cancel } = nodeService.getNodeTags(node.id.toString());

    request
      .then((res) => {
        if (isMounted) {
          setTagNodes(res.data);
        }
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching node tags:", error);
        }
      });

    return () => {
      isMounted = false;
      cancel();
    };
  }, [node.id]);

  return (
    <>
      {tagNodes.length > 0 && (
        <TableContainer>
          <Table variant="striped">
            <Thead>
              <Tr>
                <Th>
                  <Text as="b">Key</Text>
                </Th>
                <Th>
                  <Text as="b">Value</Text>
                </Th>
              </Tr>
            </Thead>
            <Tbody>
              {tagNodes.map((tagNode) => (
                <Tr key={tagNode.id}>
                  <Td>
                    <Text>{tagNode.properties.map.key}</Text>
                  </Td>
                  <Td>{tagNode.properties.map.value}</Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </TableContainer>
      )}
    </>
  );
};

export default TagPanel;
