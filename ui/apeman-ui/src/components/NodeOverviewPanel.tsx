import { Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";

import { Node } from "../services/nodeService";
import AccountOverviewPanel from "./AccountOverviewPanel";
import RoleOverviewPanel from "./RoleOverviewPanel";
import ResourceOverview from "./ResourceOverview";
import NodeOverview from "./NodeOverview";
import StatementOverview from "./StatementOverview";
import { kinds } from "../services/nodeService";
import PolicyOverview from "./PolicyOverview";
import UserOverviewPanel from "./UserOverviewPanel";
import GroupOverviewPanel from "./GroupOverviewPanel";
<<<<<<< HEAD
=======
import ActionOverviewPanel from "./ActionOverviewPanel";
>>>>>>> main

interface Props {
  node: Node;
}

const NodeOverviewPanel = ({ node }: Props) => {
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
    [kinds.AWSAction, "Action Overview"],
  ]);

  return (
    <>
      <Tabs
        width="100%"
        height="100%"
        isFitted
        variant="soft-rounded"
        size="sm"
      >
        <TabList>
<<<<<<< HEAD
          {nodeKinds.map((kind) => (
            <Tab fontSize="xs" key={kind}>
              {tabTitleMap.get(kind)}
            </Tab>
          ))}
=======
          {nodeKinds.map(
            (kind) =>
              tabTitleMap.get(kind) && (
                <Tab fontSize="xs" key={kind}>
                  {tabTitleMap.get(kind)}
                </Tab>
              )
          )}
>>>>>>> main
          <Tab fontSize="xs" key="nodeOverview">
            Node Overview
          </Tab>
        </TabList>
        <TabPanels>
<<<<<<< HEAD
          {nodeKinds.map((kind) => (
            <TabPanel key={kind}>
              {kind === kinds.AWSAccount ? (
                <AccountOverviewPanel node={node}></AccountOverviewPanel>
              ) : null}
              {kind === kinds.AWSRole ? (
                <RoleOverviewPanel node={node}></RoleOverviewPanel>
              ) : null}
              {kind === kinds.AWSUser ? (
                <UserOverviewPanel node={node}></UserOverviewPanel>
              ) : null}
              {kind === kinds.AWSGroup ? (
                <GroupOverviewPanel node={node}></GroupOverviewPanel>
              ) : null}
              {kind === kinds.UniqueArn ? (
                <ResourceOverview node={node} />
              ) : null}
              {kind === kinds.AWSStatement ? (
                <StatementOverview node={node}></StatementOverview>
              ) : null}
              {kind === kinds.AWSManagedPolicy ? (
                <PolicyOverview node={node}></PolicyOverview>
              ) : null}
            </TabPanel>
          ))}
=======
          {nodeKinds.map(
            (kind) =>
              tabTitleMap.get(kind) && (
                <TabPanel key={kind}>
                  {kind === kinds.AWSAccount ? (
                    <AccountOverviewPanel node={node}></AccountOverviewPanel>
                  ) : null}
                  {kind === kinds.AWSRole ? (
                    <RoleOverviewPanel node={node}></RoleOverviewPanel>
                  ) : null}
                  {kind === kinds.AWSUser ? (
                    <UserOverviewPanel node={node}></UserOverviewPanel>
                  ) : null}
                  {kind === kinds.AWSGroup ? (
                    <GroupOverviewPanel node={node}></GroupOverviewPanel>
                  ) : null}
                  {kind === kinds.UniqueArn ? (
                    <ResourceOverview node={node} />
                  ) : null}
                  {kind === kinds.AWSStatement ? (
                    <StatementOverview node={node}></StatementOverview>
                  ) : null}
                  {kind === kinds.AWSManagedPolicy ? (
                    <PolicyOverview node={node}></PolicyOverview>
                  ) : null}
                  {kind === kinds.AWSAction ? (
                    <ActionOverviewPanel node={node}></ActionOverviewPanel>
                  ) : null}
                </TabPanel>
              )
          )}
>>>>>>> main
          <TabPanel>
            <NodeOverview node={node}></NodeOverview>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </>
  );
};

export default NodeOverviewPanel;
