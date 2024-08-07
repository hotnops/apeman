import { useEffect, useState, FC } from "react";
import {
  ReactTree,
  TreeNode as BaseTreeNode,
  TreeNodeId,
} from "@naisutech/react-tree";
import nodeService, { Node, getNodeLabel } from "../services/nodeService";
import accountService from "../services/accountService";
import { Box } from "@chakra-ui/react";

import { BiMinus, BiPlus } from "react-icons/bi";

interface ExtendedTreeNode extends BaseTreeNode {
  type: string;
  items?: ExtendedTreeNode[];
}

interface Props {
  graphNodes: { [key: string]: Node };
  setGraphNodes: React.Dispatch<React.SetStateAction<{ [key: string]: Node }>>;
}

const NodeExplorer: FC<Props> = ({ setGraphNodes }) => {
  const [treeNodes, setTreeNodes] = useState<Map<TreeNodeId, ExtendedTreeNode>>(
    new Map()
  );

  useEffect(() => {
    accountService.getAllAccounts().request.then((resp) => {
      const newTreeNodes = new Map<TreeNodeId, ExtendedTreeNode>();
      resp.data.forEach((account: string) => {
        const item: ExtendedTreeNode = {
          id: account,
          label: account,
          parentId: null,
          type: "AWSAccount",
        };
        newTreeNodes.set(account, item);
      });
      setTreeNodes(newTreeNodes);
    });
  }, []);

  const fetchNode = async (nodeId: string) => {
    try {
      const request = nodeService.getNodeByID(nodeId);
      const res = await request.request;
      if (res.data.id !== 0) {
        setGraphNodes((prev) => ({ ...prev, [res.data.id]: res.data }));
      }
    } catch (error) {
      console.error("Failed to fetch node", error);
    }
  };

  const getChildNodes = (nodeId: TreeNodeId) => {
    const node = treeNodes.get(nodeId);
    if (!node) return;

    const createAndSetNode = (id: string, label: string, type: string) => {
      const newNode: ExtendedTreeNode = { id, label, parentId: nodeId, type };
      setTreeNodes((prev) => new Map(prev.set(id, newNode)));
    };

    if (node.type === "AWSAccount") {
      createAndSetNode(`${nodeId}-policies`, "Policies", "AWSManagedPolicy");
      createAndSetNode(`${nodeId}-roles`, "Roles", "AWSRole");
      createAndSetNode(`${nodeId}-users`, "Users", "AWSUser");
      createAndSetNode(`${nodeId}-groups`, "Groups", "AWSGroup");
    } else {
      const kind = node.type;

      const accountId = (nodeId as string).split("-")[0];
      const params = new Map<string, string>([
        ["account_id", accountId],
        ["kind", kind],
      ]);

      nodeService
        .getNodesWithParams(params)
        .request.then((resp) => {
          const items: ExtendedTreeNode[] = (resp.data as Node[]).map(
            (graphNode: Node) => ({
              id: graphNode.id,
              label: getNodeLabel(graphNode),
              parentId: nodeId,
              type: kind,
            })
          );

          const updatedNode = { ...node, items };
          setTreeNodes((prev) => new Map(prev.set(nodeId, updatedNode)));
        })
        .catch((error) => {
          console.error("Failed to fetch child nodes", error);
        });
    }
  };

  return (
    <Box>
      <ReactTree
        messages={{
          emptyItems: "No items to display",
          loading: "Loading...",
          noData: "No data to display",
        }}
        truncateLongText
        nodes={[...treeNodes.values()]}
        enableIndicatorAnimations
        enableItemAnimations
        onToggleOpenNodes={(nodes: TreeNodeId[]) => {
          nodes.forEach((node) => getChildNodes(node));
        }}
        onToggleSelectedNodes={(nodes: TreeNodeId[]) => {
          nodes.forEach((node) => {
            const parsed = Number(node as string);
            if (!isNaN(parsed)) {
              fetchNode(node as string);
            }
          });
        }}
        RenderIcon={({ open, type }) => {
          if (type == "node") {
            if (open) {
              return <BiMinus style={{ transform: "rotate(90deg)" }} />;
            } else {
              return <BiPlus />;
            }
          } else {
            return <></>;
          }
        }}
      />
    </Box>
  );
};

export default NodeExplorer;
