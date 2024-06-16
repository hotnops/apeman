import {
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Button,
  HStack,
  Table,
  Tbody,
  Td,
  Text,
  Tr,
} from "@chakra-ui/react";

import { PiGraph } from "react-icons/pi";
import { Path } from "../services/pathService";

interface Props {
  name: string;
  paths: Path[];
  pathLabelFunction: (p: Path) => string;
  pathFunction?: (p: Path) => void;
}

const AccordionList = ({
  name,
  paths,
  pathLabelFunction,
  pathFunction,
}: Props) => {
  if (paths === null) {
    return <>No paths found</>;
  }
  return (
    <AccordionItem width="100%">
      <HStack width="100%" justifyContent="left">
        <AccordionButton width="100%" justifyContent="space-between">
          <Text width="80%" textAlign="left">
            {name}
          </Text>
          <Text width="10%">{paths.length}</Text>
          <AccordionIcon width="10%"></AccordionIcon>
        </AccordionButton>
        <Button
          onClick={() => {
            paths.map((path) => {
              if (pathFunction) {
                pathFunction(path);
              }
            });
          }}
        >
          <PiGraph />
        </Button>
      </HStack>
      <AccordionPanel>
        <Table overflowX="scroll">
          <Tbody>
            {paths.map((path) => (
              <Tr>
                <Td textOverflow="ellipsis" width="80%">
                  {pathLabelFunction(path)}
                </Td>

                {pathFunction && (
                  <Td width="20%">
                    <Button
                      size="xs"
                      onClick={() => {
                        if (pathFunction) {
                          pathFunction(path);
                        }
                      }}
                    >
                      <PiGraph />
                    </Button>
                  </Td>
                )}
              </Tr>
            ))}
          </Tbody>
        </Table>
      </AccordionPanel>
    </AccordionItem>
  );
};

export default AccordionList;
