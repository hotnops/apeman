import { useEffect, useState } from "react";
import AccordionList from "./AccordionList";
import nodeService from "../services/nodeService";
import { Node } from "../services/nodeService";
import {
  Path,
  addPathToGraph,
  getNodesFromPaths,
} from "../services/pathService";
import { useApemanGraph } from "../hooks/useApemanGraph";

interface Props {
  account_id?: string;
}

const IndirectTierZeroList = ({ account_id = "" }: Props) => {
  const [paths, setPaths] = useState<Path[]>([]);
  const { addNode, addEdge } = useApemanGraph();

  function graphPath(n: Node) {
    // Get Path for node
    console.log(`Node id: ${n.id}`);
    const path = paths.filter((path) => path.Nodes[0].id == n.id)[0];
    console.log("PATH");
    console.log(path);
    addPathToGraph(path, addNode, addEdge);
  }

  useEffect(() => {
    const { request, cancel } = nodeService.getTierZeroPaths(account_id);
    request.then((res) => {
      setPaths(res.data);
    });
    return cancel;
  }, []);
  return (
    <>
      {paths.length == 0 ? (
        <AccordionList
          nodes={getNodesFromPaths(paths).filter(
            (node) => !("tier_zero" in node.properties.map)
          )}
          name="Indirect Tier Zero Principals"
          pathFunction={graphPath}
        ></AccordionList>
      ) : (
        <p>No paths found</p>
      )}
    </>
  );
};

export default IndirectTierZeroList;
