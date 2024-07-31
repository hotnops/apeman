import React, { useState } from "react";

import { Node } from "../services/nodeService";
import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Skeleton,
  Text,
} from "@chakra-ui/react";

import ActionPathList from "./ActionPathList";
import { GetRoleRSOPActions } from "../services/roleService";
import { GetUserRSOPActions } from "../services/userService";

interface Props {
  node: Node;
}

const RSOPPanel = ({ node }: Props) => {
  const [rsopActionMap, setRsopActionMap] = useState({});
  const [isRsopLoaded, setIsRsopLoaded] = useState(false);

  const fetchRsopActionMap = async () => {
    var res;
    if (node.kinds.includes("AWSRole")) {
      const { request } = GetRoleRSOPActions(node.properties.map.roleid);
      res = await request;
    } else if (node.kinds.includes("AWSUser")) {
      const { request } = GetUserRSOPActions(node.properties.map.userid);
      res = await request;
    } else {
      return;
    }

    setRsopActionMap(res.data);
    setIsRsopLoaded(true);
  };

  const handleAccordionChange = (index: number | number[]) => {
    if (!Array.isArray(index)) {
      index = [index];
    }

    if (index.length != 0 && !isRsopLoaded) {
      fetchRsopActionMap();
    }
  };
  return (
    <Accordion
      allowMultiple={true}
      width="100%"
      onChange={handleAccordionChange}
    >
      <AccordionItem>
        <AccordionButton>
          <Box flex="1" textAlign="left">
            <Text as="b">Resultant Set of Policy</Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
        <AccordionPanel pb={4}>
          {isRsopLoaded ? (
            <ActionPathList actionPathMap={rsopActionMap} />
          ) : (
            <Skeleton height="200px" />
          )}
        </AccordionPanel>
      </AccordionItem>
    </Accordion>
  );
};

export default RSOPPanel;
