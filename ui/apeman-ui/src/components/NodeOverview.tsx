import { Node } from "../services/nodeService";
import { Table, Tbody, Td, Tr, Text } from "@chakra-ui/react";
import TagPanel from "./TagPanel";

interface Props {
  node: Node;
}

const NodeOverview = ({ node }: Props) => {
  return (
    <>
      <Table>
        <Tbody>
          {Object.keys(node.properties.map).map((key) => (
            <Tr>
              <Td>{key}</Td>
              <Td>{node.properties.map[key]}</Td>
            </Tr>
          ))}
          <Tr>
            <Td>Kinds</Td>
            <Td>
              {node.kinds.map((kind) => (
                <p>{kind}</p>
              ))}
            </Td>
          </Tr>
        </Tbody>
      </Table>
      <Text fontSize="xl">Tags</Text>
      <TagPanel></TagPanel>
    </>
  );
};

export default NodeOverview;
