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
  const [principals, setPrincipals] = useState<string[]>([]);

  const fetchData = () => {
    const controller = new AbortController();

    const request = apiClient.get<string[]>(endpoint, {
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
        setPrincipals(res.data);
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
      {/* <Accordion width="100%" allowMultiple={true}>
        {principals.map((path) => (
          <PermissionItem
            name={path}
            actions={[]}
            resourceId={resourceId()}
          ></PermissionItem>
        ))}
      </Accordion> */}
    </>
  );
};

export default PermissionList;
