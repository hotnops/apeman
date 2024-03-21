import apiClient from "./api-client";

export function getActiveConditionKeys() {
  const controller = new AbortController();
  const request = apiClient.get("/conditionkeys/active", {
    signal: controller.signal,
  });
  return {
    request,
    cancel: () => {
      controller.abort();
    },
  };
}
