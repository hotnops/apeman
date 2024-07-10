import { Node } from "../services/nodeService";
import { Table, Tbody, Td, Tr, Text } from "@chakra-ui/react";
import TagPanel from "./TagPanel";

interface Props {
  node: Node;
}

const NodeOverview = ({ node }: Props) => {
  return (
    <>
      <Table size="sm">
        <Tbody>
          {Object.keys(node.properties.map).map((key) => (
            <Tr>
              <Td>
                <b>{key}</b>
              </Td>
              <Td textAlign="right">{node.properties.map[key]}</Td>
            </Tr>
          ))}
          <Tr>
            <Td>
              <b>kinds</b>
            </Td>
            <Td textAlign="right">
              {node.kinds.map((kind) => (
                <p>{kind}</p>
              ))}
            </Td>
          </Tr>
          <Tr>
            <Td>
              <b>node id</b>
            </Td>
            <Td textAlign="right">{node.id}</Td>
          </Tr>
        </Tbody>
      </Table>
      <Text fontSize="md">Tags</Text>
      <TagPanel></TagPanel>
    </>
  );
};

export default NodeOverview;
