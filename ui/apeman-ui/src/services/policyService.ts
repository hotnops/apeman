import apiClient from "./api-client";
import { Node } from "./nodeService";

const MANAGED_BASE_PATH = "/managedpolicies";
const INLINE_BASE_PATH = "/inlinepolicies";



class PolicyService {

    getInlinePolicyJson(policyId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(INLINE_BASE_PATH + "/" + policyId + "/generatepolicy", {
        signal: controller.signal,
      });
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

    getPolicyPrincipalNodes(policyId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(MANAGED_BASE_PATH + "/" + policyId + "/principals", {
        signal: controller.signal,
      });
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

    getManagedPolicyJSON(policyId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(MANAGED_BASE_PATH + "/" + policyId + "/generatepolicy", {
        signal: controller.signal,
      });
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

    getInlinePolicyJSON(policyHash: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(INLINE_BASE_PATH + "/" + policyHash + "/generatepolicy", {
        signal: controller.signal,
      });
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

    getInlinePolicyNode(prinicpalNode: Node) {
      const controller = new AbortController();

      var request
  
      if (prinicpalNode.kinds.includes("AWSUser")) {
        request = apiClient.get<Node>("/users/" + prinicpalNode.properties.map.userid + "/inlinepolicy", {
          signal: controller.signal,
        });
      } else if (prinicpalNode.kinds.includes("AWSRole")) {
        request = apiClient.get<Node>("/roles/" + prinicpalNode.properties.map.roleid + "/inlinepolicy", {
          signal: controller.signal,
        });
      } else if (prinicpalNode.kinds.includes("AWSGroup")) {
        request = apiClient.get<Node>("/groups/" + prinicpalNode.properties.map.groupid + "/inlinepolicy", {
          signal: controller.signal,
        });
      }
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

    getNodesAttachedToPolicy(policyId: string, kind: string) {
      const controller = new AbortController();

      var request
  
      if (kind === "managed") {
        request = apiClient.get(MANAGED_BASE_PATH + "/" + policyId + "/nodes", {
          signal: controller.signal,
        });
      }
      else {
        request = apiClient.get(INLINE_BASE_PATH + "/" + policyId + "/nodes", {
          signal: controller.signal,
        });
      }

      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }

  }
  
  export default new PolicyService();