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

export function GetRoleRSOP(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/roles/${roleId}/rsop`, {
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
  getRoleManagedPolicyNodes(roleid: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + roleid + "/managedpolicies", {
      signal: controller.signal,
    });

    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getRoleInlinePolicyNodes(roleid: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + roleid + "/inlinepolicies", {
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
