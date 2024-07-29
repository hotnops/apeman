import { useEffect } from "react";
import apiClient from "../services/api-client";
import { Input } from "@chakra-ui/react";
import { Node } from "../services/nodeService";

interface Props {
  search: string;
  setSearch: (s: string) => void;
  setSearchResults: (nodes: Node[]) => void;
  variant?: string;
  placeholder?: string;
  kinds?: string[];
}

const SearchBar = ({
  search,
  setSearch,
  setSearchResults,
  variant = "unstyled",
  placeholder = "",
}: Props) => {
  const handleChange = (e: any) => {
    setSearch(e.target.value);
  };

  useEffect(() => {
    if (search.length > 4) {
      const controller = new AbortController();
      const signal = controller.signal;
      const request = apiClient.get(`/search?searchQuery=${search}`, {
        signal,
      });

      request
        .then((res: any) => {
          var newNodes: Node[] = [];
          Object.keys(res.data).map((item) => {
            newNodes.push(res.data[item]);
          });
          setSearchResults(newNodes);
        })
        .catch((error) => {
          if (error.code === "ERR_CANCELED") {
            console.log("Request was aborted");
          } else {
            console.error("An error occurred:", error);
          }
        });

      return () => {
        controller.abort();
      };
    } else {
      setSearchResults([]);
    }
  }, [search]);

  return (
    <Input
      onChange={handleChange}
      value={search}
      height="100%"
      margin="5px"
      variant={variant}
      placeholder={placeholder}
    />
  );
};

export default SearchBar;
