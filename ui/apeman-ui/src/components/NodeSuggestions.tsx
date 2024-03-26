import { Node } from "../services/nodeService";
import { List, ListItem, Text, useTheme } from "@chakra-ui/react";
import NodeListItem from "./NodeListItem";

interface Props {
  nodes: Node[];
  searchQuery: string;
  onItemSelect: (node: Node) => void;
}

const NodeSuggestions = ({ nodes, onItemSelect }: Props) => {
  var truncated = false;
  if (nodes.length > 5) {
    truncated = true;
  }

  const theme = useTheme();

  return (
    <List
      boxShadow="lg"
      borderBottomRadius={15}
      width="100%"
      backgroundColor={theme.colors.white}
    >
      {nodes.slice(0, 5).map((node) => (
        <ListItem
          border={10}
          _hover={{ bg: theme.colors.gray[50] }}
          cursor="pointer"
          onClick={() => onItemSelect(node)}
        >
          <NodeListItem node={node} />
        </ListItem>
      ))}
      {truncated && (
        <ListItem width="100%">
          <Text textAlign="center" fontSize="large">
            ...
          </Text>
        </ListItem>
      )}
    </List>
  );
};

export default NodeSuggestions;
