import { Card, HStack } from "@chakra-ui/react";
import { SearchIcon } from "@chakra-ui/icons";
import { useEffect, useRef, useState } from "react";
import { Node } from "../services/nodeService";
import NodeSuggestions from "./NodeSuggestions";
import SearchBar from "./SearchBar";
import { useApemanGraph } from "../hooks/useApemanGraph";

const NavBar = () => {
  const [searchResults, setSearchResults] = useState<Node[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
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
        position="relative"
        top="15px"
        left="10px"
        width="25em"
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
        </HStack>
        {showSuggestions && searchResults.length > 0 ? (
          <NodeSuggestions
            nodes={searchResults}
            searchQuery={search}
            onItemSelect={(node) => {
              console.log("Selected node" + node);
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
