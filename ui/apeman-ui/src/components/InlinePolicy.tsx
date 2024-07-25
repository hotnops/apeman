import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Text,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";

import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Node } from "../services/nodeService";
import { AsyncGetInlinePolicyJSON } from "../services/policyService";
import nodeService from "../services/nodeService";
import roleService from "../services/roleService";
import userService from "../services/userService";
import groupService from "../services/groupService";

interface Props {
  principalNode: Node;
}

const InlinePolicy = ({ principalNode }: Props) => {
  const [inlinePolicies, setInlinePolicies] = useState<string[]>([]);

  useEffect(() => {
    const fetchInlinePolicies = async () => {
      try {
        let res;
        if (principalNode.kinds.includes("AWSRole")) {
          const { request } = roleService.getRoleInlinePolicyNodes(
            principalNode.properties.map.roleid
          );
          res = await request;
        } else if (principalNode.kinds.includes("AWSUser")) {
          const { request } = userService.getUserInlinePolicyNodes(
            principalNode.properties.map.userid
          );
          res = await request;
        } else if (principalNode.kinds.includes("AWSGroup")) {
          const { request } = groupService.getGroupInlinePolicyNodes(
            principalNode.properties.map.groupid
          );
          res = await request;
        }

        const policyRequests = res?.data.nodes.map((node: Node) => {
          const { request } = nodeService.getNodeByID(node.id.toString());
          return request;
        });

        const responses = await Promise.all(policyRequests);

        const policiesSet = new Set<string>();

        await Promise.all(
          responses.map(async (res) => {
            const policy = await AsyncGetInlinePolicyJSON(
              (res.data as Node).properties.map["hash"]
            );
            policiesSet.add(policy as string);
          })
        );

        setInlinePolicies(Array.from(policiesSet));
      } catch (error) {
        console.error("Error fetching policies:", error);
      }
    };

    fetchInlinePolicies();
  }, []);

  return (
    <Accordion allowMultiple={true} width="100%">
      <AccordionItem>
        <AccordionButton>
          <Box as="span" flex="1" textAlign="left">
            <Text as="b" fontSize="sm">
              Inline Policies
            </Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
        <AccordionPanel>
          {inlinePolicies.map((policy) => (
            <SyntaxHighlighter language="json" style={coy}>
              {policy && JSON.stringify(policy, null, 4)}
            </SyntaxHighlighter>
          ))}
        </AccordionPanel>
      </AccordionItem>
    </Accordion>
  );
};

export default InlinePolicy;
