import React, { useEffect, useState } from "react";
import {
  ReactTree,
  TreeNode,
  TreeNodeId,
  useReactTreeApi,
} from "@naisutech/react-tree";
import nodeService, { Node, getNodeLabel } from "../services/nodeService";
import accountService from "../services/accountService";
import { Box } from "@chakra-ui/react";

const awsAccountsToNode = (accountNames: string[]) => {
  var returnItems: {}[] = [];
  var index = 0;
  accountNames.map((accountName) => {
    returnItems.push({
      id: accountName,
      label: accountName,
      parentId: null,
      type: "AWSAccount",
    });

    returnItems.push({
      id: `${accountName}-policies`,
      label: "Policies",
      parentId: accountName,
      type: "AWSManagedPolicy",
    });

    returnItems.push({
      id: `${accountName}-roles`,
      label: "Roles",
      parentId: accountName,
      type: "AWSRole",
    });

    returnItems.push({
      id: `${accountName}-users`,
      label: "Users",
      parentId: accountName,
      type: "AWSUser",
    });

    returnItems.push({
      id: `${accountName}-groups`,
      label: "Groups",
      parentId: accountName,
      type: "AWSGroup",
    });
  });

  return returnItems;
};

interface Props {
  graphNodes: { [key: string]: Node };
  setGraphNodes: (nodes: { [key: string]: Node }) => void;
}

const NodeExplorer = ({ graphNodes, setGraphNodes }: Props) => {
  const [treeNodes, setTreeNodes] = useState<Map<TreeNodeId, TreeNode>>(
    new Map()
  );

  useEffect(() => {
    accountService.getAllAccounts().request.then((resp) => {
      resp.data.map((account) => {
        const item = {
          id: account,
          label: account,
          parentId: null,
          type: "AWSAccount",
        };

        setTreeNodes((prev) => new Map(prev.set(account, item)));
      });
    });
  }, []);

  const fetchNode = async (nodeId: string) => {
    const request = nodeService.getNodeByID(nodeId.toString());
    const res = await request.request;
    if (res.data.id != 0) {
      setGraphNodes((graphNodes) => {
        const newNodes = { ...graphNodes };
        newNodes[res.data.id] = res.data;
        return newNodes;
      });
    }
  };

  const getChildNodes = (nodeId: TreeNodeId) => {
    console.log("Getting child nodes");
    const node = treeNodes.get(nodeId);
    var kind = "";
    if (!node) {
      return;
    }
    console.log(node);
    if (node.type === "AWSAccount") {
      const polNode = {
        id: `${nodeId}-policies`,
        label: "Policies",
        parentId: nodeId,
        type: "AWSManagedPolicy",
      };

      setTreeNodes((prev) => new Map(prev.set(`${nodeId}-policies`, polNode)));

      const roleNode = {
        id: `${nodeId}-roles`,
        label: "Roles",
        parentId: nodeId,
        type: "AWSRole",
      };

      setTreeNodes((prev) => new Map(prev.set(`${nodeId}-roles`, roleNode)));

      const userNode = {
        id: `${nodeId}-users`,
        label: "Users",
        parentId: nodeId,
        type: "AWSUser",
      };

      setTreeNodes((prev) => new Map(prev.set(`${nodeId}-users`, userNode)));

      const groupNode = {
        id: `${nodeId}-groups`,
        label: "Groups",
        parentId: nodeId,
        type: "AWSGroup",
      };

      setTreeNodes((prev) => new Map(prev.set(`${nodeId}-groups`, groupNode)));
    } else {
      if (node.type == "AWSRole") {
        kind = "AWSRole";
      } else if (node.type == "AWSManagedPolicy") {
        kind = "AWSManagedPolicy";
      } else if (node.type == "AWSUser") {
        kind = "AWSUser";
      } else if (node.type == "AWSGroup") {
        kind = "AWSGroup";
      }

      const account_id = (nodeId as string).split("-")[0];
      const params = new Map<string, string>();
      params.set("account_id", account_id);
      params.set("kind", kind);
      var items: TreeNode[] = [];
      nodeService.getNodesWithParams(params).request.then((resp) => {
        resp.data.map((graphNode) => {
          console.log(graphNode);
          items.push({
            id: graphNode.id,
            label: getNodeLabel(graphNode),
            parentId: nodeId,
            type: kind,
          });
        });
        // Add items to the TreeNode

        var typeRoot = treeNodes.get(node.id);
        typeRoot.items = items;

        setTreeNodes((prev) => new Map(prev.set(nodeId, typeRoot)));
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
          nodes.map((node) => getChildNodes(node));
        }}
        onToggleSelectedNodes={(nodes: TreeNodeId[]) => {
          console.log("Selected nodes changed");

          nodes.map((node) => {
            fetchNode(node as string);
          });
        }}
      ></ReactTree>
    </Box>
  );
};

export default NodeExplorer;
