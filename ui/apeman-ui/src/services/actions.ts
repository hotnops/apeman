import apiClient from "./api-client";
import { Path } from "./pathService";

export function GetActionPolicies(actionName: string) {
  const controller = new AbortController();
  const request = apiClient.get<Path[]>(`/actions/${actionName}/policies`, {
    signal: controller.signal,
  });

  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}