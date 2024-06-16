import apiClient from "./api-client";

const BASE_PATH = "/roles";

export function GetInboundRoles(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/roles/${roleId}/inboundroles`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

export function GetOutboundRoles(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/roles/${roleId}/outboundroles`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

export function GetInboundPaths(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/roles/${roleId}/inboundpaths`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

class RoleService {
  getRolePolicyNodes(roleid: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + roleid + "/policies", {
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

export default new RoleService();
