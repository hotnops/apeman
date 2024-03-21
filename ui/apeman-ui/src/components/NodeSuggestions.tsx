import React from "react";
import { Node } from "../services/nodeService";
import {
  Card,
  HStack,
  List,
  ListItem,
  Text,
  Image,
  Box,
  useTheme,
} from "@chakra-ui/react";
import NodeListItem from "./NodeListItem";

interface Props {
  nodes: Node[];
  searchQuery: string;
  onItemSelect: (node: Node) => void;
}

const getSearchTermFromNode = (node: Node, searchQuery: string): string => {
  if ("arn" in node.properties.map) {
    const haystack = node.properties.map["arn"];
    if (haystack.toLowerCase().includes(searchQuery.toLowerCase()))
      return haystack;
  }
  if ("name" in node.properties.map) {
    const haystack = node.properties.map["name"];
    if (haystack.toLowerCase().includes(searchQuery.toLowerCase()))
      return haystack;
  }
  if ("hash" in node.properties.map) {
    const haystack = node.properties.map["hash"].toLowerCase();
    if (haystack.toLowerCase().includes(searchQuery.toLowerCase()))
      return haystack;
  }
  return "";
};

const NodeSuggestions = ({ nodes, searchQuery, onItemSelect }: Props) => {
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
