import apiClient from "./api-client";

const BASE_PATH = "/groups";

class GroupService {

    getGroupMembershipPaths(groupId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(BASE_PATH + "/" + groupId + "/members", {
        signal: controller.signal,
      });
  
      return {
        request,
        cancel: () => {
          controller.abort();
        },
      };
    }
  
    getGroupInlinePolicyNode(groupId: string) {
      const controller = new AbortController();
  
      const request = apiClient.get(BASE_PATH + "/" + groupId + "/inlinepolicy", {
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
  