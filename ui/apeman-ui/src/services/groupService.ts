import apiClient from "./api-client";

const BASE_PATH = "/groups";

class GroupService {
  
    getGroupInlinePolicyNodes(groupId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(BASE_PATH + "/" + groupId + "/inlinepolicies", {
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
  
  export default new GroupService();
  