import { Box, HStack, Image, Text } from "@chakra-ui/react";
import { Node, getIconURL, getNodeLabel } from "../services/nodeService";

interface Props {
  node: Node;
}

const NodeListItem = ({ node }: Props) => {
  return (
    <HStack>
      <Box boxSize="25px" flexShrink={0}>
        <Image
          src={getIconURL(node.kinds)}
          objectFit="fill"
          boxSize="100%"
          padding="2px"
          marginLeft="5px"
        ></Image>
      </Box>
      <Text
        fontSize="sm"
        overflow="hidden"
        whiteSpace="nowrap"
        textOverflow="ellipsis"
        padding={2}
        margin={1}
      >
        {getNodeLabel(node)}
      </Text>
    </HStack>
  );
};

export default NodeListItem;
