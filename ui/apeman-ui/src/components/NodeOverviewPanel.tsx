import { Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";

import { Node } from "../services/nodeService";
import AccountOverviewPanel from "./AccountOverviewPanel";
import RoleOverviewPanel from "./RoleOverviewPanel";
import ResourceOverview from "./ResourceOverview";
import NodeOverview from "./NodeOverview";
import StatementOverview from "./StatementOverview";
import { kinds } from "../services/nodeService";
import TagPanel from "./TagPanel";

interface Props {
  node: Node;
}

const NodeOverviewPanel = ({ node }: Props) => {
  console.log("Rendering NodeOverviewPanel");
  const nodeKinds = node.kinds;

  let tabTitleMap = new Map<string, string>([
    [kinds.AWSAccount, "Account Overview"],
    [kinds.AWSRole, "Role Overview"],
    [kinds.AWSUser, "User Overview"],
    [kinds.AWSManagedPolicy, "Policy Overview"],
    [kinds.AWSInlinePolicy, "Policy Overview"],
    [kinds.AWSGroup, "Group Overview"],
    [kinds.UniqueArn, "Resource Overview"],
    [kinds.AWSStatement, "Statement Overview"],
  ]);

  return (
    <>
      <Tabs width="100%">
        <TabList>
          {nodeKinds.map((kind) => (
            <Tab key={kind}>{tabTitleMap.get(kind)}</Tab>
          ))}
          <Tab key="nodeOverview">Node Overview</Tab>
        </TabList>
        <TabPanels>
          {nodeKinds.map((kind) => (
            <TabPanel key={kind}>
              {kind === kinds.AWSAccount ? (
                <AccountOverviewPanel node={node}></AccountOverviewPanel>
              ) : null}
              {kind === kinds.AWSRole ? (
                <RoleOverviewPanel node={node}></RoleOverviewPanel>
              ) : null}
              {kind === kinds.UniqueArn ? (
                <ResourceOverview node={node} />
              ) : null}
              {kind === kinds.AWSStatement ? (
                <StatementOverview node={node}></StatementOverview>
              ) : null}
            </TabPanel>
          ))}
          <TabPanel>
            <NodeOverview node={node}></NodeOverview>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </>
  );
};

export default NodeOverviewPanel;
