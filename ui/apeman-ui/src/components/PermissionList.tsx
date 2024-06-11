import {
  Button,
  HStack,
  IconButton,
  Input,
  List,
  ListItem,
  Text,
} from "@chakra-ui/react";
import { MdFilterAlt } from "react-icons/md";
import { useEffect, useState } from "react";
import apiClient from "../services/api-client";
import { IoCloseCircle } from "react-icons/io5";
import { useApemanGraph } from "../hooks/useApemanGraph";
import { Path, addPathToGraph } from "../services/pathService";

interface Props {
  children: string;
  endpoint: string;
}

const PermissionList = ({ children, endpoint }: Props) => {
  const [_, setError] = useState<Error | null>(null);
  const [__, setLoading] = useState(false);
  const [, setSearchQuery] = useState("");
  const [showFilter, setShowFilter] = useState(false);
  const [paths, setPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.log("text change");
    setSearchQuery(event.target.value);
  };

  const fetchData = () => {
    const controller = new AbortController();

    const request = apiClient.get<Path[]>(endpoint, {
      signal: controller.signal,
    });

    return {
      request,
      abort: () => {
        controller.abort();
      },
    };
  };

  useEffect(() => {
    const { request, abort } = fetchData();
    setError(null);
    setLoading(true);
    request
      .then((res) => {
        //setPermissions(res.data);
        // Create a string to action list map
        setPaths(res.data);
      })
      .catch((err) => {
        setError(err);
      })
      .finally(() => {
        setLoading(false);
      });
    return abort;
  }, []);

  return (
    <>
      {showFilter && (
        <HStack>
          <Input onChange={handleInputChange}></Input>
          <IconButton
            aria-label="Filter actions"
            icon={<IoCloseCircle />}
            size="25px"
            onClick={() => {
              setShowFilter(false);
              setSearchQuery("");
            }}
          ></IconButton>
        </HStack>
      )}
      <HStack justifyContent="space-between" paddingY={2}>
        <h1>
          <Text fontWeight="bold">{children}</Text>
        </h1>
        <IconButton
          aria-label="Expand Filter bar"
          icon={<MdFilterAlt />}
          onClick={() => setShowFilter(true)}
        ></IconButton>
      </HStack>

      <List>
        {paths.map((path) => (
          <ListItem>
            <HStack>
              <Text>
                {path.Nodes[path.Nodes.length - 1].properties.map["name"]}
              </Text>
              <Button
                onClick={() => {
                  addPathToGraph(path, addNode, addEdge);
                }}
              >
                Add To Graph
              </Button>
            </HStack>
          </ListItem>
        ))}
      </List>
    </>
  );
};

export default PermissionList;
