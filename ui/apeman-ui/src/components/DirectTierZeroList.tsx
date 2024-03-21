import React, { useEffect, useState } from "react";
import AccordionList from "./AccordionList";
import nodeService from "../services/nodeService";
import { Node } from "../services/nodeService";

interface Props {
  account_id?: string;
}

const DirectTierZeroList = ({ account_id = "" }: Props) => {
  const [nodes, setNodes] = useState<Node[]>([]);
  useEffect(() => {
    const { request, cancel } = nodeService.getTierZeroNodes(account_id);
    request.then((res) => {
      setNodes(res.data);
    });
    return cancel;
  }, []);
  return (
    <AccordionList
      nodes={nodes}
      name="Direct Tier Zero Principals"
    ></AccordionList>
  );
};

export default DirectTierZeroList;
