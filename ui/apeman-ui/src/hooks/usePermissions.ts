import { useEffect, useState } from "react";
import apiClient from "../services/api-client";

export interface Permission {
  arn: string;
  actions: string[];
  pathIndex: number;
}

export interface Node {
  id: string;
  label: string;
}

interface Property {
  map: { account_id: string };
  deleted: {};
  modified: {};
}

interface NodeResponse {
  id: number;
  kinds: string[];
  properties: Property;
}

export const useGetNodes = (
  endpoint: string,
  setNodes: (nodes: Node[]) => void,
  setLoading: (b: boolean) => void
) => {
  const nodes: Node[] = [];

  useEffect(() => {
    apiClient
      .get<NodeResponse[]>(endpoint)
      .then((res) => {
        res.data.map((node: NodeResponse) => {
          nodes.push({
            id: node.id.toString(),
            label: node.properties.map.account_id,
          });
        });
        setNodes(nodes);
      })
      .catch(() => {
        setNodes([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);
};

const usePermissions = (endpoint: string, refreshCount: number) => {
  const [error, setError] = useState("");
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [isLoading, setLoading] = useState(false);

  const controller = new AbortController();

  useEffect(() => {
    setLoading(true);

    apiClient
      .get<Permission[]>(endpoint, { signal: controller.signal })
      .then((res) => {
        setPermissions(res.data);
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
    return () => {
      controller.abort();
    };
  }, [refreshCount]);

  return { permissions, error, isLoading, controller };
};

export default usePermissions;
