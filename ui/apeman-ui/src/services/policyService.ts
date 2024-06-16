import apiClient from "./api-client";

const BASE_PATH = "/managedpolicies";

class PolicyService {
    getPolicyPrincipalNodes(policyId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(BASE_PATH + "/" + policyId + "/principals", {
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