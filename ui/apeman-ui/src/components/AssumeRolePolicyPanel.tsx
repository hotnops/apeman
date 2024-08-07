import { useEffect, useState } from "react";
import { Node } from "../services/nodeService";
import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Text,
} from "@chakra-ui/react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";
import roleService from "../services/roleService";

interface Props {
  roleNode: Node;
}

const AssumeRolePolicyPanel = ({ roleNode }: Props) => {
  const [assumeRolePolicies, setAssumeRolePolicies] = useState<string[]>([]);

  useEffect(() => {
    const { request, cancel } = roleService.getAssumeRolePolicyObject(
      roleNode.properties.map.roleid
    );

    request
      .then((res) => {
        console.log(res.data);
        setAssumeRolePolicies(res.data.Statement);
      })
      .catch((error) => {
        console.error("Error fetching assume role policy:", error);
      });

    return () => {
      cancel();
    };
  }, [roleNode]);

  return (
    <Accordion allowMultiple={true} width="100%">
      <AccordionItem>
        <AccordionButton>
          <Box as="span" flex="1" textAlign="left">
            <Text as="b" fontSize="sm">
              Assume Role Policy
            </Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
        <AccordionPanel>
          {assumeRolePolicies.map((policy, index) => (
            <SyntaxHighlighter key={index} language="json" style={coy}>
              {policy && JSON.stringify(policy, null, 4)}
            </SyntaxHighlighter>
          ))}
        </AccordionPanel>
      </AccordionItem>
    </Accordion>
  );
};

export default AssumeRolePolicyPanel;
