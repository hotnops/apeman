import {
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  List,
  Text,
} from "@chakra-ui/react";
import { Path } from "../services/pathService";

interface Props {
  path: Path;
}
const PermissionItem = ({ path }: Props) => {
  return (
    <AccordionItem>
      <h2>
        <AccordionButton>
          <Box as="span" flex="1" textAlign="left">
            <Text fontWeight="bold">{path.Nodes[0].properties.map["arn"]}</Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
      </h2>
      <AccordionPanel>
        <List spacing={3}></List>
      </AccordionPanel>
    </AccordionItem>
  );
};

export default PermissionItem;
