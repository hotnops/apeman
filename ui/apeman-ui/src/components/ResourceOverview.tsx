import { Card } from "@chakra-ui/react";
import PermissionList from "./PermissionList";
import { encode } from "url-safe-base64";
import { Buffer } from "buffer";
import { Node } from "../services/nodeService";

interface Props {
  node: Node;
}

const ResourceOverview = ({ node }: Props) => {
  const b64encode = (str: string): string =>
    Buffer.from(str, "binary").toString("base64");

  const resourceArn = node.properties.map.arn;
  const endpoint =
    "resources/" + encode(b64encode(resourceArn)) + "/inboundpermissions";
  return (
    <>
      <PermissionList endpoint={endpoint}>InboundPermissions</PermissionList>
    </>
  );
};

export default ResourceOverview;
