import {
  Editable,
  EditableInput,
  EditablePreview,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@chakra-ui/react";
import React, { useEffect, useState } from "react";
import { getActiveConditionKeys } from "../services/conditionService";
import { Node } from "../services/nodeService";

const ContextSettings = () => {
  const [conditionKeys, setConditionKeys] = useState<string[]>([]);

  useEffect(() => {
    const { request, cancel } = getActiveConditionKeys();
    var keyList: string[] = [];
    request.then((res) => {
      res.data.map((node: Node) => {
        keyList.push(node.properties.map.name);
      });
      setConditionKeys(keyList);
    });
    console.log(keyList);

    return cancel;
  }, []);
  return (
    <Table>
      <Thead>
        <Tr>
          <Th>Condition Key</Th>
          <Th>Condition Value</Th>
        </Tr>
      </Thead>
      <Tbody>
        {conditionKeys.map((key) => (
          <Tr key={key}>
            <Td>{key}</Td>
            <Td>
              <Editable defaultValue="Unset">
                <EditablePreview />
                <EditableInput />
              </Editable>
            </Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
};

export default ContextSettings;
