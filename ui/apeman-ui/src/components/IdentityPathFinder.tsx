import { useState } from "react";
import { Node } from "../services/nodeService";
import { useTheme } from "@emotion/react";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { Box, HStack } from "@chakra-ui/react";
import { MdOutlinePinDrop, MdTripOrigin } from "react-icons/md";
import NodeListItem from "./NodeListItem";
import HoverIcon from "./HoverIcon";
import { IoCloseCircleOutline, IoEllipsisVerticalSharp } from "react-icons/io5";
import { IoMdArrowRoundDown } from "react-icons/io";
import SearchBar from "./SearchBar";
import NodeSuggestions from "./NodeSuggestions";

interface Props {
  nodes: Node[];
  setPathNodes: (updateFn: (prevNodes: Node[]) => Node[]) => void;
}

const IdentityPathFinder = ({ nodes, setPathNodes }: Props) => {
  const theme = useTheme();
  const [searchResults, setSearchResults] = useState<Node[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const { addNode } = useApemanGraph();
  return (
    <>
      {nodes.map((node) => (
        <>
          <HStack margin="5px" height="2em" justifyContent="space-between">
            <Box
              boxSize="25px"
              flexShrink={0}
              margin="5px"
              justifyContent="center"
              alignItems="center"
              display="flex"
            >
              <MdTripOrigin size="75%"></MdTripOrigin>
            </Box>
            <Box overflow="hidden" textOverflow="ellipses">
              <NodeListItem node={node}></NodeListItem>
            </Box>
            {nodes.length > 0 && (
              <Box boxSize="25px" flexShrink={0}>
                <HoverIcon
                  iconColor={(theme as any).colors.gray[500]}
                  hoverColor={(theme as any).colors.gray[900]}
                >
                  <IoCloseCircleOutline
                    onClick={() => {
                      setPathNodes(() => []);
                      setSearchQuery("");
                    }}
                    size="100%"
                  ></IoCloseCircleOutline>
                </HoverIcon>
              </Box>
            )}
          </HStack>
          <HStack margin="5px" height="1em" justifyContent="space-between">
            <Box
              boxSize="25px"
              flexShrink={0}
              margin="5px"
              justifyContent="center"
              alignItems="center"
              display="flex"
            >
              <IoEllipsisVerticalSharp></IoEllipsisVerticalSharp>
            </Box>
            <IoMdArrowRoundDown></IoMdArrowRoundDown>
            <Box visibility="hidden"></Box>
          </HStack>
        </>
      ))}
      <HStack margin="5px" height="2em">
        <Box
          boxSize="25px"
          flexShrink={0}
          margin="5px"
          justifyContent="center"
          alignItems="center"
          display="flex"
        >
          <MdOutlinePinDrop size="75%"></MdOutlinePinDrop>
        </Box>

        <SearchBar
          search={searchQuery}
          setSearch={setSearchQuery}
          setSearchResults={setSearchResults}
          variant="outline"
          placeholder="Enter search term or click on node"
        ></SearchBar>
      </HStack>

      <NodeSuggestions
        nodes={searchResults}
        searchQuery={searchQuery}
        onItemSelect={(node: Node) => {
          addNode(node);
          setPathNodes((prev: Node[]) => [...prev, node]);
          setSearchQuery("");
          setSearchResults([]);
        }}
      ></NodeSuggestions>
    </>
  );
};

export default IdentityPathFinder;
