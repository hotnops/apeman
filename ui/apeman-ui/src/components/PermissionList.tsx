import { Accordion, HStack, IconButton, Input, Text } from "@chakra-ui/react";
import { MdFilterAlt } from "react-icons/md";
import { useEffect, useState } from "react";
import PermissionItem from "./PermissionItem";
import apiClient from "../services/api-client";
import { IoCloseCircle } from "react-icons/io5";

type PrincipalToActionMap = {
  [key: string]: string[];
};

interface Props {
  children: string;
  endpoint: string;
  resourceId: () => number;
}

const PermissionList = ({ children, endpoint, resourceId }: Props) => {
  const [_, setError] = useState<Error | null>(null);
  const [__, setLoading] = useState(false);
  const [showFilter, setShowFilter] = useState(false);
  const [paths, setPaths] = useState<PrincipalToActionMap>({});
  const [searchQuery, setSearchQuery] = useState("");

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.log("text change");
    setSearchQuery(event.target.value);
  };

  function filterMapBySubstring(
    map: PrincipalToActionMap,
    query: string
  ): PrincipalToActionMap {
    const filteredMap: PrincipalToActionMap = {};

    if (query === "") {
      return map;
    }

    for (const [key, value] of Object.entries(map)) {
      if (value.some((str) => str.includes(query))) {
        filteredMap[key] = value;
      }
    }

    return filteredMap;
  }

  const fetchData = () => {
    const controller = new AbortController();

    const request = apiClient.get<PrincipalToActionMap>(endpoint, {
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
        console.log("RSOP");
        console.log(res.data);
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
          <Text fontWeight="bold" fontSize="sm">
            {children}
          </Text>
        </h1>
        <IconButton
          aria-label="Expand Filter bar"
          icon={<MdFilterAlt />}
          onClick={() => setShowFilter(true)}
        ></IconButton>
      </HStack>
      <Accordion width="100%" allowMultiple={true}>
        {Object.entries(filterMapBySubstring(paths, searchQuery)).map(
          ([key, values]) => (
            <PermissionItem
              name={key}
              actions={values}
              resourceId={resourceId()}
            ></PermissionItem>
          )
        )}
      </Accordion>
    </>
  );
};

export default PermissionList;
