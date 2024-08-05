import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Button,
  HStack,
  Text,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";

import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Node } from "../services/nodeService";
import { PiGraph } from "react-icons/pi";
import { useApemanGraph } from "../hooks/useApemanGraph";
import policyService from "../services/policyService";
import { Path, addPathToGraph } from "../services/pathService";

interface Props {
  node: Node;
}

const InlinePolicy = ({ node }: Props) => {
  const [inlineStatements, setInlineStatements] = useState<string[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  const graphPolicyNodes = (node: Node) => {
    console.log(node);
    const { request } = policyService.getNodesAttachedToPolicy(
      node.properties.map.hash,
      "inline"
    );

    request
      .then((res) => {
        console.log(res.data);
        res.data.map((path: Path) => addPathToGraph(path, addNode, addEdge));
      })
      .catch((error) => {
        if (error.code !== "ERR_CANCELED") {
          console.error("Error fetching node permissions:", error);
        }
      });
  };

  useEffect(() => {
    const { request, cancel } = policyService.getInlinePolicyJSON(
      node.properties.map.hash
    );

    request
      .then((res) => {
        setInlineStatements(res.data.Statement);
      })
      .catch((error) => {
        console.error("Error fetching inline policy:", error);
      });

    return () => {
      cancel();
    };
  }, [node]);

  return (
    <Accordion allowMultiple={true} width="100%">
      <AccordionItem>
        <HStack>
          <AccordionButton>
            <Box as="span" flex="1" textAlign="left">
              <Text as="b" fontSize="sm">
                Inline Policy
              </Text>
            </Box>
            <AccordionIcon />
          </AccordionButton>
          <Button onClick={() => graphPolicyNodes(node)}>
            <PiGraph />
          </Button>
        </HStack>

        <AccordionPanel>
          {inlineStatements.map((policy) => (
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
