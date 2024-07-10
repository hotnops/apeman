import apiClient from "./api-client";

const MANAGED_BASE_PATH = "/managedpolicies";
const INLINE_BASE_PATH = "/inlinepolicies";

export const AsyncGetInlinePolicyJSON = (policyHash: string) => {
  return apiClient.get(INLINE_BASE_PATH + "/" + policyHash + "/generatepolicy").then((res) => {
    return res.data as string;
  });
}


class PolicyService {
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

  }
  
  export default new PolicyService();