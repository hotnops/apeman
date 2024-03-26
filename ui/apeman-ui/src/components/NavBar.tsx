import { Box, Card, HStack, useTheme } from "@chakra-ui/react";
import { SearchIcon } from "@chakra-ui/icons";
import { useEffect, useRef, useState } from "react";
import { Node } from "../services/nodeService";
import NodeSuggestions from "./NodeSuggestions";
import { RiDirectionLine } from "react-icons/ri";
import HoverIcon from "./HoverIcon";
import SearchBar from "./SearchBar";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  closeNavBar: () => void;
}

const NavBar = ({ closeNavBar }: Props) => {
  const [searchResults, setSearchResults] = useState<Node[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const theme = useTheme();
  const [search, setSearch] = useState("");
  const handleFocusChange = (event: any) => {
    if (wrapperRef.current != null) {
      if (!wrapperRef.current.contains(event.target)) {
        setShowSuggestions(false);
      } else {
        setShowSuggestions(true);
      }
    }
  };
  useEffect(() => {
    document.addEventListener("mousedown", handleFocusChange);
    return () => {
      document.removeEventListener("mousedown", handleFocusChange);
    };
  });
  const { addNode, setActiveElement: setActiveNode } = useApemanGraph();

  return (
    <div ref={wrapperRef}>
      <Card
        position="fixed"
        top="15px"
        left="10px"
        width="30em"
        height="35px"
        zIndex={1}
        margin="5px"
        borderTopRadius={15}
        borderBottomRadius={
          searchResults.length !== 0 && showSuggestions ? 0 : 20
        }
      >
        <HStack height="100%">
          <SearchIcon margin="10px"></SearchIcon>
          <SearchBar
            search={search}
            setSearch={setSearch}
            setSearchResults={setSearchResults}
          ></SearchBar>
          <Box
            marginRight={7}
            onClick={closeNavBar}
            boxSize="25px"
            flexShrink={0}
          >
            <HoverIcon
              iconColor={theme.colors.gray[500]}
              hoverColor={theme.colors.gray[900]}
            >
              <RiDirectionLine style={{ width: "100%", height: "100%" }} />
            </HoverIcon>
          </Box>
        </HStack>
        {showSuggestions && searchResults.length > 0 ? (
          <NodeSuggestions
            nodes={searchResults}
            searchQuery={search}
            onItemSelect={(node) => {
              addNode(node);
              setActiveNode(node);
              setSearch("");
              setSearchResults([]);
            }}
          ></NodeSuggestions>
        ) : null}
      </Card>
    </div>
  );
};

export default NavBar;
