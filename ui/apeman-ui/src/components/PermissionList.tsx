import {
  Accordion,
  Button,
  HStack,
  IconButton,
  Input,
  Spinner,
  Text,
} from "@chakra-ui/react";
import { MdFilterAlt } from "react-icons/md";
import { Permission } from "../hooks/usePermissions";
import { useEffect, useRef, useState } from "react";
import PermissionItem from "./PermissionItem";
import apiClient from "../services/api-client";
import { IoCloseCircle } from "react-icons/io5";
import { Icon } from "reagraph";

interface Props {
  children: string;
  endpoint: string;
}

const PermissionList = ({ children, endpoint }: Props) => {
  const [error, setError] = useState<Error | null>(null);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [isLoading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [showFilter, setShowFilter] = useState(false);

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.log("text change");
    setSearchQuery(event.target.value);
  };

  const getFilteredPermissions = (permissions: Permission[]) => {
    console.log("Filtering list");
    return permissions.filter((permission) =>
      Object.keys(permission.actions).some((action) =>
        action.includes(searchQuery)
      )
    );
  };

  const fetchData = () => {
    const controller = new AbortController();

    const request = apiClient.get<Permission[]>(endpoint, {
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
        setPermissions(res.data);
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

      <Accordion allowToggle>
        {getFilteredPermissions(permissions).map((permission) => (
          <PermissionItem
            key={permission.arn}
            queryString={searchQuery}
            {...permission}
          ></PermissionItem>
        ))}
      </Accordion>
    </>
  );
};

export default PermissionList;
