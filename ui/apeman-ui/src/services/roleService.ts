import apiClient from "./api-client";
import { NodeResponse } from "./nodeService";
import { Path } from "./pathService";

const BASE_PATH = "/roles";

export function GetInboundRoles(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get<Path[]>(`/roles/${roleId}/inboundroles`, {
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
  const request = apiClient.get<Path[]>(`/roles/${roleId}/outboundroles`, {
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

export function GetRoleRSOPActions(roleId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/roles/${roleId}/rsop/actions`, {
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

    const request = apiClient.get<NodeResponse>(BASE_PATH + "/" + roleid + "/managedpolicies", {
      signal: controller.signal,
    });

    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getRoleInlinePolicyNode(roleid: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + roleid + "/inlinepolicy", {
      signal: controller.signal,
    });

    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getAssumeRolePolicyObject(roleid: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + roleid + "/generateassumerolepolicy", {
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
