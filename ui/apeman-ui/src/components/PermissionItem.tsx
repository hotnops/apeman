import {
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Box,
  Highlight,
  List,
  ListItem,
  Text,
} from "@chakra-ui/react";

interface Props {
  arn: string;
  actions: {};
  queryString: string;
}
const PermissionItem = ({ arn, actions, queryString }: Props) => {
  return (
    <AccordionItem>
      <h2>
        <AccordionButton>
          <Box as="span" flex="1" textAlign="left">
            <Text fontWeight="bold">{arn}</Text>
          </Box>
          <AccordionIcon />
        </AccordionButton>
      </h2>
      <AccordionPanel>
        <List spacing={3}>
          {Object.keys(actions).map((action) => (
            <ListItem>
              {queryString ? (
                <Highlight query={queryString} styles={{ fontWeight: "bold" }}>
                  {action}
                </Highlight>
              ) : (
                <>{action}</>
              )}
            </ListItem>
          ))}
        </List>
      </AccordionPanel>
    </AccordionItem>
  );
};

export default PermissionItem;
