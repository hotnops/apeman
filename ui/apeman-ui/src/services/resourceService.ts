import apiClient from "./api-client";
import { Node } from "./nodeService";
import { Buffer } from "buffer";

const BASE_PATH = "/resources";


const b64encode = (str: string): string =>
    Buffer.from(str, "binary").toString("base64");


class ResourceService {

    getResourceActions(node: Node) {
        const resourceArn = node.properties.map.arn;
        const endpoint =
            BASE_PATH + "/" + b64encode(resourceArn) + "/actions";

        const controller = new AbortController();

        const request = apiClient.get(endpoint, {
            signal: controller.signal,
        });

        return {
            request,
            cancel: () => {
                controller.abort();
            },
        };
    }

    getPrincipalsWithResourceAndAction(node: Node, action: string) {
        const resourceArn = node.properties.map.arn;
        const endpoint =
            BASE_PATH +
            "/" +
            b64encode(resourceArn) +
            "/inboundpermissions?actionName=" + action;

        const controller = new AbortController();

        const request = apiClient.get(endpoint, {
            signal: controller.signal,
        });

        return {
            request,
            cancel: () => {
                controller.abort();
            },
        };
    }

    getActionsWithResourceAndPrincipal(node: Node, principal: string) {
        const resourceArn = node.properties.map.arn;
        const endpoint =
            BASE_PATH +
            "/" +
            b64encode(resourceArn) +
            "/inboundpermissions/principals/" +
            b64encode(principal);

        const controller = new AbortController();

        const request = apiClient.get(endpoint, {
            signal: controller.signal,
        });

        return {
            request,
            cancel: () => {
                controller.abort();
            },
        };
    }

    getInboundResourcePermissions(node: Node) {
        const resourceArn = node.properties.map.arn;
        const endpoint = BASE_PATH + "/" + b64encode(resourceArn) + "/inboundpermissions/principals";

        const controller = new AbortController();

        const request = apiClient.get<string[]>(endpoint, {
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
  
  export default new ResourceService();