import apiClient from "./api-client";

const BASE_PATH = "/users";

export function GetOutboundRoles(userId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/users/${userId}/outboundroles`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

export function GetInboundPaths(userId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/users/${userId}/inboundpaths`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

export function GetUserRSOP(userId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/users/${userId}/rsop`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}


export function GetUserRSOPActions(userId: string) {
  const controller = new AbortController();
  const request = apiClient.get(`/users/${userId}/rsop/actions`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}

class UserService {
  getUserPolicyNodes(userId: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + userId + "/managedpolicies", {
      signal: controller.signal,
    });

    return {
      request,
      cancel: () => {
        controller.abort();
      },
    };
  }

  getUserInlinePolicyNodes(userId: string) {
    const controller = new AbortController();

    const request = apiClient.get(BASE_PATH + "/" + userId + "/inlinepolicies", {
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

export default new UserService();
