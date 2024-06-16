package analyze

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"

	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/graphcache"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
)

type ActionPathEntry struct {
	PrincipalID  graph.ID      `json:"principal_id"`
	PrincipalArn string        `json:"principal_arn"`
	ResourceArn  string        `json:"resource_arn"`
	Action       string        `json:"action"`
	Path         graph.Path    `json:"path"`
	Effect       string        `json:"effect"`
	Statement    *graph.Node   `json:"statement"`
	Conditions   graph.NodeSet `json:"conditions"`
}

type ActionPathSet struct {
	ActionPaths []ActionPathEntry `json:"action_paths"`
}

func (a *ActionPathEntry) IsEqual(other ActionPathEntry) bool {
	return a.PrincipalArn == other.PrincipalArn && a.Action == other.Action
}

func (a *ActionPathSet) Add(actionPath ActionPathEntry) {
	if !a.ContainsActionPath(actionPath) {
		a.ActionPaths = append(a.ActionPaths, actionPath)
	}
}

func (a *ActionPathSet) GetPrincipals() []string {
	principals := []string{}

	for _, actionPath := range a.ActionPaths {
		principals = append(principals, actionPath.PrincipalArn)
	}

	return principals
}

func (a *ActionPathSet) AddPathSet(actionPathSet *ActionPathSet) {
	for _, actionPath := range actionPathSet.ActionPaths {
		a.Add(actionPath)
	}
}

func (a *ActionPathSet) ContainsActionPath(actionPath ActionPathEntry) bool {
	for _, path := range a.ActionPaths {
		if path.IsEqual(actionPath) {
			return true
		}
	}
	return false
}

func (a *ActionPathSet) RemoveActionPathEntry(actionPath ActionPathEntry) {
	tempPaths := make([]ActionPathEntry, 0)
	for _, path := range a.ActionPaths {
		if !path.IsEqual(actionPath) {
			tempPaths = append(tempPaths, path)
		}
	}
	a.ActionPaths = tempPaths
}

func (a *ActionPathSet) GetPaths() graph.PathSet {
	paths := graph.NewPathSet()
	for _, actionPath := range a.ActionPaths {
		paths.AddPath(actionPath.Path)
	}
	return paths
}

func ResolvePaths(pathSet graph.PathSet) (graph.PathSet, error) {
	// Given a set of paths, evaluate denies and conditions and
	// return a set of paths that are allowed

	denyPathSet := graph.NewPathSet()
	allowPathSet := graph.NewPathSet()
	conditionalDenyPathSet := graph.NewPathSet()
	conditionalAllowPathSet := graph.NewPathSet()

	// For each path, determine if the statement is a deny
	// if it is, move it to the deny path set and any other
	// path that contains any actions in the deny path set
	for _, path := range pathSet {
		// Sort paths by their satement type
		statement := GetNodeFromPathByKind(path, aws.AWSStatement)
		condition := GetNodeFromPathByKind(path, aws.AWSCondition)
		if effect, err := statement.Properties.Get("effect").String(); err != nil {
			continue
		} else {
			if effect == "Deny" {
				if condition != nil {
					conditionalDenyPathSet.AddPath(path)
				} else {
					denyPathSet.AddPath(path)
				}
			} else {
				if condition != nil {
					conditionalAllowPathSet.AddPath(path)
				} else {
					allowPathSet.AddPath(path)
				}
			}
		}
	}

	// For each deny path, move any allow paths that contain the same actions
	for _, denyPath := range denyPathSet {
		action := GetNodeFromPathByKind(denyPath, aws.AWSAction)
		for i, allowPath := range allowPathSet {
			if allowPath.Terminal().ID == action.ID {
				RemovePathByIndex(&allowPathSet, i)
			}
		}
	}

	for _, denyPath := range conditionalDenyPathSet {
		if satisified, err := ResolveConditons(denyPath); err != nil {
			continue
		} else if satisified {
			action := denyPath.Terminal()
			for i, allowPath := range allowPathSet {
				if allowPath.Terminal().ID == action.ID {
					RemovePathByIndex(&allowPathSet, i)
				}
			}
		}
	}
	for _, allowPath := range conditionalAllowPathSet {
		if satisified, err := ResolveConditons(allowPath); err != nil {
			continue
		} else if satisified {
			allowPathSet.AddPath(allowPath)
		}
	}

	return allowPathSet, nil
}

func IdentityPolicyPathToActionPathSet(paths graph.PathSet) *ActionPathSet {
	actionPathSets := new(ActionPathSet)
	for _, path := range paths {
		actionPathEntry := ActionPathEntry{}
		actionPathEntry.Path = path
		actionPathEntry.Statement = GetNodeFromPathByKind(path, aws.AWSStatement)
		actionPathEntry.Effect, _ = actionPathEntry.Statement.Properties.Get("effect").String()
		actionNode := GetNodeFromPathByKind(path, aws.AWSAction)
		actionPathEntry.Action, _ = actionNode.Properties.Get("name").String()
		principalNode, _ := GetPrincipalFromIdentityPath(path)
		actionPathEntry.PrincipalID = principalNode.ID
		actionPathEntry.PrincipalArn, _ = principalNode.Properties.Get("arn").String()
		actionPathEntry.Conditions = GetNodesFromPathByKind(path, aws.AWSCondition)
		actionPathSets.Add(actionPathEntry)
	}
	return actionPathSets
}

func ResourcePolicyPathToActionPathSet(paths graph.PathSet) *ActionPathSet {
	actionPathSets := new(ActionPathSet)
	for _, path := range paths {
		actionPathEntry := ActionPathEntry{}
		actionPathEntry.Path = path
		actionPathEntry.Statement = GetNodeFromPathByKind(path, aws.AWSStatement)
		actionPathEntry.Effect, _ = actionPathEntry.Statement.Properties.Get("effect").String()
		actionNode := GetNodeFromPathByKind(path, aws.AWSAction)
		actionPathEntry.Action, _ = actionNode.Properties.Get("name").String()
		principalNode, _ := GetPrincipalFromResourcePath(path)
		actionPathEntry.PrincipalArn, _ = principalNode.Properties.Get("arn").String()
		actionPathEntry.Conditions = GetNodesFromPathByKind(path, aws.AWSCondition)
		actionPathSets.Add(actionPathEntry)
	}
	return actionPathSets
}

func (p *ActionPathEntry) String() string {
	return fmt.Sprintf("PrincipalArn: %s, Action: %s, Effect: %s", p.PrincipalArn, p.Action, p.Effect)
}

func (p *ActionPathSet) SplitByEffect() (allow *ActionPathSet, deny *ActionPathSet) {
	allow = new(ActionPathSet)
	deny = new(ActionPathSet)
	for _, actionPath := range p.ActionPaths {
		if actionPath.Effect == "Allow" {
			allow.Add(actionPath)
		} else {
			deny.Add(actionPath)
		}
	}
	return allow, deny
}

func GetPrincipalFromIdentityPath(path graph.Path) (*graph.Node, error) {
	prinNode := path.Nodes[0]
	if !prinNode.Kinds.ContainsOneOf(aws.AWSUser, aws.AWSRole) {
		return nil, fmt.Errorf("principal not found")
	}
	return prinNode, nil
}

func GetPrincipalFromResourcePath(path graph.Path) (*graph.Node, error) {
	for _, edge := range path.Edges {
		if edge.Kind == aws.Resource {
			// Get the node by ID
			resourceNode, err := GetNodeFromPathByID(path, edge.EndID)
			if err != nil {
				return nil, err
			}
			if resourceNode.Kinds.ContainsOneOf(aws.AWSResourceBlob) {
				for _, expandedEdge := range path.Edges {
					if expandedEdge.StartID == resourceNode.ID {
						principalNode, err := GetNodeFromPathByID(path, expandedEdge.EndID)
						if err != nil {
							return nil, err
						}
						return principalNode, nil
					}
				}
				return nil, fmt.Errorf("principal not found")
			}
			return resourceNode, nil
		}

	}

	return nil, fmt.Errorf("resource not found")
}

func ActionPathSetIntersection(resourceActionSet *ActionPathSet, identityActionPathSet *ActionPathSet) (*ActionPathSet, error) {
	denyPathSet := new(ActionPathSet)
	resolvedPaths := new(ActionPathSet)

	resourceAllowSet, resourceDenySet := resourceActionSet.SplitByEffect()
	identityAllowSet, identityDenySet := identityActionPathSet.SplitByEffect()

	denyPathSet.AddPathSet(resourceDenySet)
	denyPathSet.AddPathSet(identityDenySet)

	for _, denyPath := range denyPathSet.ActionPaths {
		// Remove all paths that contain this action
		//if awsconditions.ResolveConditions(denyPath.Conditions) {
		resourceAllowSet.RemoveActionPathEntry(denyPath)
		identityAllowSet.RemoveActionPathEntry(denyPath)
		//}
	}

	// Each allow path must be in both sets
	for _, identityAllowPath := range identityAllowSet.ActionPaths {
		// if the identity path is conditionally true
		if resourceAllowSet.ContainsActionPath(identityAllowPath) {
			resolvedPaths.Add(identityAllowPath)
		}
	}
	return resolvedPaths, nil

}

func GetPrincipalsOfPolicy(ctx context.Context, db graph.Database, policyNode *graph.Node) (graph.NodeSet, error) {
	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		principals    = make(graph.NodeSet)
		mapLock       = &sync.Mutex{}
	)

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: policyNode,
		Driver: func(policyNode *graph.Node) traversal.PatternContinuation {
			return traversal.NewPattern().Outbound(
				query.And(
					query.KindIn(query.End(), aws.AWSUser, aws.AWSRole, aws.AWSGroup),
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Equals(query.StartID(), policyNode.ID),
				),
			)
		}(policyNode).Do(func(terminal *graph.PathSegment) error {
			mapLock.Lock()
			defer mapLock.Unlock()
			principals.Add(terminal.Path().Terminal())
			return nil
		}),
	}); err != nil {
		return nil, err
	}
	return principals, nil
}

func GetAttachedKinds(ctx context.Context, db graph.Database, node *graph.Node, kinds graph.Kinds) (graph.NodeSet, error) {
	var (
		nodeSet = graph.NewNodeSet()
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.EndID(), node.ID),
				query.Kind(query.Relationship(), aws.AttachedTo),
				query.KindIn(query.Start(), kinds...))
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for relationship := range cursor.Chan() {
				node, err := ops.FetchNode(tx, relationship.StartID)
				if err != nil {
					return err
				}
				nodeSet.Add(node)
			}

			return nil
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return nodeSet, nil

}

func extractAccountID(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("Invalid ARN format")
	}
	accountID := parts[4]
	return accountID, nil

}

func GetDirectStatementsOnResource(resourceArn string) traversal.PatternContinuation {
	return traversal.NewPattern().Inbound(
		query.And(
			query.Kind(query.Start(), aws.AWSStatement),
			query.Equals(query.EndProperty("arn"), resourceArn),
			query.Kind(query.Relationship(), aws.Resource),
		),
	)
}

func GetIndirectStatementsOnResource(resourceArn string) traversal.PatternContinuation {
	return traversal.NewPattern().Inbound(
		query.And(
			query.Kind(query.Relationship(), aws.ExpandsTo),
			query.Equals(query.EndProperty("arn"), resourceArn),
		),
	).Inbound(
		query.And(
			query.Kind(query.Start(), aws.AWSStatement),
			query.Kind(query.End(), aws.AWSResourceBlob),
			query.Kind(query.Relationship(), aws.Resource),
		),
	)
}

func ResourceTypePattern(resourceNode *graph.Node) traversal.PatternContinuation {
	// MATCH (u) - [:TypeOf] -> (:AWSResourceType)
	return traversal.NewPattern().Outbound(
		query.And(
			query.Kind(query.Relationship(), aws.TypeOf),
			query.Equals(query.StartID(), resourceNode.ID),
		),
	)
}

func GetResourceTypeOfArn(ctx context.Context, db graph.Database, resourceArn string) (*graph.Node, error) {
	var (
		startNode     *graph.Node
		resourceType  *graph.Node
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		err           error
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		startNode, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.Equals(query.NodeProperty("arn"), resourceArn)
		}).First()
		return err
	}); err != nil {
		return nil, err
	}

	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ResourceTypePattern(startNode).Do(func(terminal *graph.PathSegment) error {
			resourceType = terminal.Path().Terminal()
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	return resourceType, nil
}

func GetEffectiveStatements(ctx context.Context, db graph.Database, statementToResourcePaths graph.PathSet, resourceType *graph.Node, targetActionName string) (graph.PathSet, error) {

	statementToActionPath := graph.NewPathSet()

	for _, path := range statementToResourcePaths {
		statementNode := path.Terminal()
		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if pathSet, err := ops.FetchPathSetByQuery(tx, fmt.Sprintf("MATCH p=(a:AWSStatement) - [:AllowAction|DenyAction|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (r:AWSResourceType) WHERE ID(a) = %d AND ID(r) = %d RETURN p", statementNode.ID, resourceType.ID)); err != nil {
				return err
			} else {
				for _, path := range pathSet {
					statementToActionPath.AddPath(path)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return statementToActionPath, nil
}

func GetAWSIdentityPolicyPermissions(ctx context.Context, db graph.Database, resourceArn string, action string) (graph.PathSet, error) {
	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		startNode     *graph.Node
		err           error
		lock          = &sync.Mutex{}
	)

	statementToResourcePaths := graph.NewPathSet()

	// Get the target node
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		startNode, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.Equals(query.NodeProperty("arn"), resourceArn)
		}).First()
		return err
	}); err != nil {
		return nil, err
	}

	accountID, _ := startNode.Properties.Get("account_id").String()

	// Get all statements that have the target resource in scope
	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: GetDirectStatementsOnResource(resourceArn).Do(func(terminal *graph.PathSegment) error {
			lock.Lock()
			defer lock.Unlock()
			if terminal.Path().Terminal().Kinds.ContainsOneOf(aws.AWSStatement) {
				statementToResourcePaths.AddPath(terminal.Path())
			}
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: GetIndirectStatementsOnResource(resourceArn).Do(func(terminal *graph.PathSegment) error {
			lock.Lock()
			defer lock.Unlock()
			if terminal.Path().Terminal().Kinds.ContainsOneOf(aws.AWSStatement) {
				statementToResourcePaths.AddPath(terminal.Path())
			}
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// At this point, statement_ids represents all of the statements that have the
	// target ARN in scope, but that doesn't mean that the statement can actually
	// act on the resource. For example, a statement that has the action iam:UpdateRolePolicy
	// and resource "*" can't act on an S3 bucket. In order to determine if a statement
	// can actually modify a resource, we need to filter out all statements that don't have
	// an action that's applicable to the target resource.
	resourceType, err := GetResourceTypeOfArn(ctx, db, resourceArn)
	if err != nil {
		return nil, err
	}

	// effectiveStatements represent all the statements where the target resource
	// is in scope AND it has at least one AWS action that can act on the resource
	returnPaths, err := GetEffectiveStatements(ctx, db, statementToResourcePaths, resourceType, action)
	if err != nil {
		return nil, err
	}

	effectiveStatements := returnPaths.AllNodes().ContainingNodeKinds(aws.AWSStatement)

	// Add all the resource paths that contain the effectiveStatements
	for _, path := range statementToResourcePaths {
		// If the path contains a statement that's not in the effectiveStatements, skip it
		if !effectiveStatements.Contains(path.Terminal()) {
			continue
		}
		returnPaths.AddPath(path)
	}

	for statementID, statement := range effectiveStatements {
		principalPaths, err := GetPrinciplesAttachedToStatement(ctx, db, statementID, accountID)
		if err != nil {
			return nil, err

		}
		returnPaths.AddPathSet(principalPaths)
		conditionPaths, err := GetStatementConditionIDs(ctx, db, statement)
		if err != nil {
			return nil, err
		}
		if len(conditionPaths) > 0 {
			returnPaths.AddPathSet(conditionPaths)
		}
	}

	return returnPaths, nil
}

func GetAWSResourceInboundPermissions(ctx context.Context, db graph.Database, resourceArn string) (graph.PathSet, error) {
	return GetAWSIdentityPolicyPermissions(ctx, db, resourceArn, "")
}

func GetRoleOutboundTrust(ctx context.Context, db graph.Database, startNode *graph.Node) (map[string]map[string]graph.NodeSet, error) {
	var (
		err            error
		lock           = &sync.Mutex{}
		traversalInst  = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		statementNodes = graph.NewNodeSet()
	)

	// Get all the statements in the assumerolepolicy document
	// that has sts:assumerole, sts:assumerolewithsaml, or sts:assumerolewithwebidentity
	// MATCH (r:AWSRole) <- [:AttachedTo] - [:AWSAssumeRolePolicy] <- [:AttachedTo] - (s:AWSStatement)
	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: func() traversal.PatternContinuation {
			return traversal.NewPattern().Inbound(
				query.And(
					query.Equals(query.EndID(), startNode.ID),
					query.Kind(query.Start(), aws.AWSAssumeRolePolicy),
					query.KindIn(query.Relationship(), aws.AttachedTo),
				),
			).Inbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Kind(query.Start(), aws.AWSStatement),
				),
			)
		}().Do(func(terminal *graph.PathSegment) error {
			lock.Lock()
			defer lock.Unlock()

			path := terminal.Path()
			statementNodes.Add(path.Terminal())
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// Map - effect/statement_hash/principals
	principals := make(map[string]map[string]graph.NodeSet)

	for _, statementNode := range statementNodes {
		// Get all directly referenced principals
		if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: statementNode,
			Driver: func() traversal.PatternContinuation {
				return traversal.NewPattern().Outbound(
					query.And(
						query.Equals(query.StartID(), statementNode.ID),
						query.KindIn(query.End(), aws.UniqueArn),
						query.Kind(query.Relationship(), aws.Resource),
					),
				)
			}().Do(func(terminal *graph.PathSegment) error {
				lock.Lock()
				defer lock.Unlock()

				effect, err := statementNode.Properties.Get("effect").String()
				if err != nil {
					return err
				}

				hash, err := statementNode.Properties.Get("hash").String()
				if err != nil {
					return err
				}

				statementMap, ok := principals[effect]
				if !ok {
					statementMap = make(map[string]graph.NodeSet)
				}

				principalNodeset, ok := statementMap[hash]
				if !ok {
					principalNodeset = graph.NewNodeSet()
				}
				path := terminal.Path()
				principalNodeset.Add(path.Terminal())
				statementMap[hash] = principalNodeset
				principals[effect] = statementMap
				return nil
			}),
		}); err != nil {
			return nil, err
		}

		// Get all indirectly referenced principals
		if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: statementNode,
			Driver: func() traversal.PatternContinuation {
				return traversal.NewPattern().Outbound(
					query.And(
						query.Equals(query.StartID(), statementNode.ID),
						query.Kind(query.End(), aws.AWSResourceBlob),
						query.Kind(query.Relationship(), aws.Resource),
					),
				).Outbound(
					query.And(
						query.Kind(query.Relationship(), aws.ExpandsTo),
						query.KindIn(query.End(), aws.UniqueArn),
					),
				)
			}().Do(func(terminal *graph.PathSegment) error {
				lock.Lock()
				defer lock.Unlock()

				path := terminal.Path()

				effect, err := statementNode.Properties.Get("effect").String()
				if err != nil {
					return err
				}

				hash, err := statementNode.Properties.Get("hash").String()
				if err != nil {
					return err
				}

				statementMap, ok := principals[effect]
				if !ok {
					statementMap = make(map[string]graph.NodeSet)
				}

				principalNodeset, ok := statementMap[hash]
				if !ok {
					principalNodeset = graph.NewNodeSet()
				}
				principalNodeset.Add(path.Terminal())
				statementMap[hash] = principalNodeset
				principals[effect] = statementMap
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	return principals, nil

}

func GetStatementConditionIDs(ctx context.Context, db graph.Database, statementNode *graph.Node) (graph.PathSet, error) {
	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		conditions    = graph.NewPathSet()
	)
	err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: statementNode,
		Driver: func() traversal.PatternContinuation {
			return traversal.NewPattern().Inbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Kind(query.Start(), aws.AWSCondition),
					query.Equals(query.EndID(), statementNode.ID),
				),
			)
		}().Do(func(terminal *graph.PathSegment) error {
			conditions.AddPath(terminal.Path())
			return nil
		}),
	})
	if err != nil {
		return nil, err
	}

	return conditions, nil
}

func GetPrinciplesAttachedToStatement(ctx context.Context, db graph.Database, statementID graph.ID, accountID string) (graph.PathSet, error) {
	var (
		err            error
		principalPaths = graph.NewPathSet()
		sliceLock      = &sync.Mutex{}
		traversalInst  = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
	)
	// MATCH  (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo] -> (:AWSInlinePolicy) - [:AttachedTo] - > (:AWSRole|AWSUser|AWSGroup) OR
	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(statementID, graph.NewProperties()),
		Driver: func() traversal.PatternContinuation {
			return traversal.NewPattern().Outbound(
				query.And(
					query.Kind(query.End(), aws.AWSPolicyDocument),
					query.Equals(query.StartID(), statementID),
					query.KindIn(query.Relationship(), aws.AttachedTo),
				),
			).Outbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Kind(query.End(), aws.AWSInlinePolicy),
				),
			).Outbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.KindIn(query.End(), aws.AWSUser, aws.AWSGroup, aws.AWSRole),
					//query.Or(
					//	query.Equals(query.EndProperty("account_id"), accountID),
					// Some targets, like S3 buckets, don't have an account
					//query.Equals("", accountID),
					//),
				),
			)
		}().Do(func(terminal *graph.PathSegment) error {
			sliceLock.Lock()
			defer sliceLock.Unlock()

			principalPaths = append(principalPaths, terminal.Path())
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// MATCH  (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo] -> (:AWSPolicyVersion) - [:AttachedTo] - > (:AWSManagedPolicy) - [:AttachedTo] -> (:AWSRole|AWSUser|AWSGroup)
	if err = traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(statementID, graph.NewProperties()),
		Driver: func() traversal.PatternContinuation {
			return traversal.NewPattern().Outbound(
				query.And(
					query.Kind(query.End(), aws.AWSPolicyDocument),
					query.Equals(query.StartID(), statementID),
					query.KindIn(query.Relationship(), aws.AttachedTo),
				),
			).Outbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Kind(query.End(), aws.AWSPolicyVersion),
				),
			).Outbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.Kind(query.End(), aws.AWSManagedPolicy),
				),
			).Outbound(
				query.And(
					query.Kind(query.Relationship(), aws.AttachedTo),
					query.KindIn(query.End(), aws.AWSUser, aws.AWSGroup, aws.AWSRole),
					query.Or(
						query.Equals(query.EndProperty("account_id"), accountID),
						// Some targets, like S3 buckets, don't have an account
						//query.Equals("", accountID),
					),
				),
			)
		}().Do(func(terminal *graph.PathSegment) error {
			sliceLock.Lock()
			defer sliceLock.Unlock()

			principalPaths = append(principalPaths, terminal.Path())
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	return principalPaths, nil

}

func GetAWSNodeByGraphID(ctx context.Context, db graph.Database, id graph.ID) (*graph.Node, error) {
	var node *graph.Node
	var err error

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		node, err = ops.FetchNode(tx, id)
		return nil
	})

	return node, err
}

func GetAWSSelfModifyingRolesAndStatements(ctx context.Context, db graph.Database) (map[graph.ID][]graph.ID, error) {
	/* Returns a map of role IDs and all the statments that are self modifying
	agains that role */
	var (
		traversalInst        = traversal.New(db, runtime.NumCPU())
		statementsActOnRoles = map[graph.ID][]graph.ID{}
		rolesToStatements    = map[graph.ID][]graph.ID{}
		mapLock              = &sync.Mutex{}
	)

	statements, err := GetAWSNodesByType(ctx, db, aws.AWSStatement)
	if err != nil {
		return nil, err
	}

	err = statements.Each(func(value uint32) (bool, error) {
		statementId := graph.ID(value)
		return true, traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: graph.NewNode(statementId, graph.NewProperties(), aws.AWSStatement),
			Driver: traversal.LightweightDriver(
				graph.DirectionOutbound,
				graphcache.New(),

				// Expand relationships that match this criteria
				query.KindIn(query.Relationship(), aws.Resource, aws.ExpandsTo),

				func(next *graph.PathSegment) bool {
					// Segment filters are called for each expansion
					//
					// For example: (a:Label1)-[r:Label10]->(b:Label2)
					//
					// The start node of the traversal is the root above: a
					//
					// The segment filter is then called for each expansion. In the first call to this function the
					// resulting path segment would contain:
					//
					// PathSegment {
					//	 Node:                b,            // The next node.
					//	 Edge: 				  r,            // The relationship that attaches to the node.
					//	 Trunk:              PathSegment {  // The previous path segment if there is one. The first segment will have a
					//      Node:    a,
					//		Edge:  nil,
					//		Trunk: nil,
					//	 },
					// }
					//
					// In this example we're looking for a path that has a terminal node with the aws.AWSRole kind. When
					// we find a node that contains that kind this will return false and halt further traversal of the
					// path.
					//return !next.Node.Kinds.ContainsOneOf(aws.AWSRole)
					return true

				},
				func(next *graph.PathSegment) {
					// This function is called when a path is considered complete and has no more expansions

					path := next.Path()
					awsRole := path.Terminal()

					// Several paths could have been completed that terminated before reaching an aws.AWSRole labeled
					// node. We have to check the terminal one last time before continuing.
					if !awsRole.Kinds.ContainsOneOf(aws.AWSRole) {
						// Exit if this isn't an aws.AWSRole labeled terminal
						return
					}

					mapLock.Lock()
					defer mapLock.Unlock()

					// Index by role
					if existingStatementIDs, hasExisting := statementsActOnRoles[statementId]; hasExisting {
						existingStatementIDs = append(existingStatementIDs, awsRole.ID)
						statementsActOnRoles[statementId] = existingStatementIDs
					} else {
						newStatementIDSlice := make([]graph.ID, 0)
						newStatementIDSlice = append(newStatementIDSlice, awsRole.ID)
						statementsActOnRoles[statementId] = newStatementIDSlice
					}
				},
			),
		})
	})

	if err != nil {
		return nil, err
	}

	// A map of statements and the roles that they act on, either
	for statementId, roleIDs := range statementsActOnRoles {
		traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: graph.NewNode(statementId, graph.NewProperties(), aws.AWSStatement),
			Driver: traversal.LightweightDriver(
				graph.DirectionOutbound,
				graphcache.New(),
				query.Kind(query.Relationship(), aws.AttachedTo),

				func(next *graph.PathSegment) bool {
					return true
				},
				func(next *graph.PathSegment) {
					path := next.Path()
					awsRole := path.Terminal()

					if roleIdMatch := func() bool {
						for _, v := range roleIDs {
							if v == awsRole.ID {
								return true
							}
						}
						return false
					}(); !roleIdMatch {
						return
					}

					mapLock.Lock()
					defer mapLock.Unlock()

					if existingStatementIDs, hasExisting := rolesToStatements[awsRole.ID]; hasExisting {
						existingStatementIDs = append(existingStatementIDs, statementId)
						rolesToStatements[awsRole.ID] = existingStatementIDs
					} else {
						newStatementSlice := make([]graph.ID, 0)
						newStatementSlice = append(newStatementSlice, statementId)
						rolesToStatements[awsRole.ID] = newStatementSlice
					}
				},
			),
		})
	}

	return rolesToStatements, nil
}

func GetAWSRoles(tx graph.Transaction) (cardinality.Duplex[uint32], error) {
	fetchedRoles, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Kind(query.Node(), aws.AWSRole)))
	if err != nil {
		return nil, err
	}
	return cardinality.NodeSetToDuplex(fetchedRoles), err
}

/*func GetStatementWithResourceTypeInScope(ctx context.Context, db graph.Database) (map[graph.ID]cardinality.Duplex[uint32], error) {
	if statementIDs, err := GetAWSNodesByType(ctx, db, aws.AWSStatement); err != nil {
		return nil, err
	}

	var (
		traversalInst = traversal.New(db, runtime.NumCPU())
		resultMap = map[graph.ID]cardinality.Duplex[uint32]{}
	)

	return resultMap, statementIDs.Each(func(value uint32) (bool, error) {
	})

}*/

func GetAWSNodesByType(ctx context.Context, db graph.Database, nodeKind graph.Kind) (cardinality.Duplex[uint32], error) {
	var fetchedNodes cardinality.Duplex[uint32]

	return fetchedNodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if nodes, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Kind(query.Node(), nodeKind))); err != nil {
			return err
		} else {
			fetchedNodes = cardinality.NodeSetToDuplex(nodes)
		}

		return nil
	})
}

func GetStatementsWithAttachRolePolicy(ctx context.Context, db graph.Database) error {
	return nil
}

func GetAWSRolesAndAttachedPolicyDocuments(ctx context.Context, db graph.Database) (map[graph.ID]cardinality.Duplex[uint32], error) {
	if policyDocumentIDs, err := GetAWSNodesByType(ctx, db, aws.AWSPolicyDocument); err != nil {
		return nil, err
	} else {
		var (
			traversalInst = traversal.New(db, runtime.NumCPU())
			resultMap     = map[graph.ID]cardinality.Duplex[uint32]{}
			mapLock       = &sync.Mutex{}
		)

		return resultMap, policyDocumentIDs.Each(func(value uint32) (bool, error) {
			// Cast the uint32 back to a graph ID
			policyDocumentID := graph.ID(value)

			// This function BreadthFirst will execute in parallel so anything happening inside of it must be thread-safe
			return true, traversalInst.BreadthFirst(ctx, traversal.Plan{
				Root: graph.NewNode(policyDocumentID, graph.NewProperties(), aws.AWSPolicyDocument),
				Driver: traversal.LightweightDriver(
					graph.DirectionOutbound,
					graphcache.New(),
					query.And(
						query.Kind(query.Start(), aws.AWSPolicyDocument),
						query.Kind(query.Relationship(), aws.AttachedTo),
						query.Kind(query.End(), aws.AWSRole),
					),
					func(next *graph.PathSegment) bool {
						// If this function returns false the driver stops chasing path expansion
						// If this function returns true the driver will go back to the database and ask for the next expansion

						// If you want to limit on depth you can return the following condition
						return (next.Depth() == 3 || next.Depth() == 2)

						// Accept unlimited depth of this traversal
						//return true
					},
					func(next *graph.PathSegment) {
						// This function is called when a path is considered complete and has no more expansions
						path := next.Path()
						awsRole := path.Terminal()

						mapLock.Lock()
						defer mapLock.Unlock()

						// Index by role
						if existingPolicyDocumentIDs, hasExisting := resultMap[awsRole.ID]; hasExisting {
							existingPolicyDocumentIDs.Add(policyDocumentID.Uint32())
						} else {
							newPolicyDocumentIDBitmap := cardinality.NewBitmap32()
							newPolicyDocumentIDBitmap.Add(policyDocumentID.Uint32())
							resultMap[awsRole.ID] = newPolicyDocumentIDBitmap
						}
					},
				),
			})
		})
	}
}

func statementIncludesAction(ctx context.Context, db graph.Database, statemetId graph.ID, action string) (bool, error) {
	var (
		traversalInst = traversal.New(db, runtime.NumCPU())
	)

	actionFound := false

	return actionFound, traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(statemetId, graph.NewProperties(), aws.AWSStatement),
		Driver: traversal.LightweightDriver(
			graph.DirectionOutbound,
			graphcache.New(),
			query.And(
				query.Kind(query.End(), aws.AWSAction),
				query.KindIn(query.Relationship(), aws.AllowAction, aws.ExpandsTo),
			),
			func(next *graph.PathSegment) bool {
				return next.Depth() < 3
			},
			func(next *graph.PathSegment) {
				// This function is called when a path is considered complete and has no more expansions
				path := next.Path()
				if len(path.Nodes) > 2 {
					actionFound = true
				}
			},
		),
	})
}

func GetSelfContainedTierZeroRoles(ctx context.Context, db graph.Database) (map[graph.ID][][]graph.ID, error) {

	selfModifyingRoles, err := GetAWSSelfModifyingRolesAndStatements(ctx, db)
	if err != nil {
		log.Printf("[!] Error getting self modified roles")
	}

	selfContainedTierZeroRoles := make(map[graph.ID][][]graph.ID)

	for roleId, statements := range selfModifyingRoles {
		attachRolePolicies := make([]graph.ID, 0)
		detachRolePolicies := make([]graph.ID, 0)
		for _, statement := range statements {
			if hasAttachRole, err := statementIncludesAction(ctx, db, statement, "iam:attachrolepolicy"); err != nil {
				return nil, err
			} else if hasAttachRole {
				attachRolePolicies = append(attachRolePolicies, statement)
			}
			if hasDetachRole, err := statementIncludesAction(ctx, db, statement, "iam:detachrolepolicy"); err != nil {
				return nil, err
			} else if hasDetachRole {
				detachRolePolicies = append(detachRolePolicies, statement)
			}
		}
		if len(attachRolePolicies) > 0 && len(detachRolePolicies) > 0 {
			selfContainedTierZeroRoles[roleId] = append(selfContainedTierZeroRoles[roleId], attachRolePolicies)
			selfContainedTierZeroRoles[roleId] = append(selfContainedTierZeroRoles[roleId], detachRolePolicies)
		}

	}

	return selfContainedTierZeroRoles, nil
}
