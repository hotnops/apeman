package analyze

import (
	"context"
	"fmt"
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

func GetAccountIDFromArn(arn string) string {
	arnParts := strings.Split(arn, ":")
	if len(arnParts) > 4 {
		return arnParts[4]
	}
	return ""
}

func ResolveAssumeRolePaths(assumeRoleSet *ActionPathSet, identityActionSet *ActionPathSet) (*ActionPathSet, error) {
	denyPathSet := new(ActionPathSet)
	condDenyPathSet := new(ActionPathSet)
	resolvedPaths := new(ActionPathSet)

	resourceAllow, resourceDeny, resourceCondAllow, resourceCondDeny := assumeRoleSet.SplitByConditionalEffect()
	identityAllow, identityDeny, identityCondAllow, identityCondDeny := identityActionSet.SplitByConditionalEffect()

	denyPathSet.AddPathSet(*resourceDeny)
	denyPathSet.AddPathSet(*identityDeny)

	for _, denyPath := range *denyPathSet {
		resourceAllow.RemoveActionPathEntry(denyPath)
		resourceCondAllow.RemoveActionPathEntry(denyPath)
		resourceCondDeny.RemoveActionPathEntry(denyPath)

		identityAllow.RemoveActionPathEntry(denyPath)
		identityCondAllow.RemoveActionPathEntry(denyPath)
		identityCondDeny.RemoveActionPathEntry(denyPath)
	}

	condDenyPathSet.AddPathSet(*resourceCondDeny)
	condDenyPathSet.AddPathSet(*identityCondDeny)

	for _, condDenyPath := range *condDenyPathSet {
		// Check if the condition is satisfied
		if resolved, err := ResolveConditions(condDenyPath); err != nil {
			return nil, err
		} else if resolved {
			resourceAllow.RemoveActionPathEntry(condDenyPath)
			resourceCondAllow.RemoveActionPathEntry(condDenyPath)

			identityAllow.RemoveActionPathEntry(condDenyPath)
			identityCondAllow.RemoveActionPathEntry(condDenyPath)
		}
	}

	for _, condAllowPath := range *resourceCondAllow {
		// Check if the condition is satisfied
		if resolved, err := ResolveConditions(condAllowPath); err != nil {
			continue
		} else if resolved {
			resourceAllow.Add(condAllowPath)
		}
	}

	for _, condAllowPath := range *identityCondAllow {
		// Check if the condition is satisfied
		if resolved, err := ResolveConditions(condAllowPath); err != nil {
			continue
		} else if resolved {
			identityAllow.Add(condAllowPath)
		}
	}

	// Each allow path must be in both sets
	for _, resourceAllowPath := range *resourceAllow {
		principalAccountId := GetAccountIDFromArn(resourceAllowPath.PrincipalArn)
		resourceAccountId := GetAccountIDFromArn(resourceAllowPath.ResourceArn)
		// If the resource specifically calls out a principal in the same account,
		// the identity policy is not needed
		if (principalAccountId == resourceAccountId) && (resourceAllowPath.IsPrincipalDirect) {
			resolvedPaths.Add(resourceAllowPath)
		} else if identityAllow.ContainsActionPath(resourceAllowPath) {
			// Check if they are in the same account and directly referenced
			resolvedPaths.Add(resourceAllowPath)
		}
	}

	return resolvedPaths, nil
}

func ResolveResourceAgainstIdentityPolicies(resourceActionSet *ActionPathSet, identityActionPathSet *ActionPathSet) (*ActionPathSet, error) {
	denyPathSet := new(ActionPathSet)
	condDenyPathSet := new(ActionPathSet)
	condAllowPathSet := new(ActionPathSet)
	resolvedPaths := new(ActionPathSet)

	resourceAllow, resourceDeny, resourceCondAllow, resourceCondDeny := resourceActionSet.SplitByConditionalEffect()
	identityAllow, identityDeny, identityCondAllow, identityCondDeny := identityActionPathSet.SplitByConditionalEffect()

	denyPathSet.AddPathSet(*resourceDeny)
	denyPathSet.AddPathSet(*identityDeny)

	for _, denyPath := range *denyPathSet {
		resourceAllow.RemoveActionPathEntry(denyPath)
		resourceCondAllow.RemoveActionPathEntry(denyPath)
		resourceCondDeny.RemoveActionPathEntry(denyPath)

		identityAllow.RemoveActionPathEntry(denyPath)
		identityCondAllow.RemoveActionPathEntry(denyPath)
		identityCondDeny.RemoveActionPathEntry(denyPath)
	}

	condDenyPathSet.AddPathSet(*resourceCondDeny)
	condDenyPathSet.AddPathSet(*identityCondDeny)

	for _, condDenyPath := range *condDenyPathSet {
		// Check if the condition is satisfied
		if resolved, err := ResolveConditions(condDenyPath); err != nil {
			return nil, err
		} else if resolved {
			resourceAllow.RemoveActionPathEntry(condDenyPath)
			resourceCondAllow.RemoveActionPathEntry(condDenyPath)

			identityAllow.RemoveActionPathEntry(condDenyPath)
			identityCondAllow.RemoveActionPathEntry(condDenyPath)
		}
	}

	condAllowPathSet.AddPathSet(*resourceCondAllow)
	condAllowPathSet.AddPathSet(*identityCondAllow)

	for _, condAllowPath := range *condAllowPathSet {
		// Check if the condition is satisfied
		if resolved, err := ResolveConditions(condAllowPath); err != nil {
			continue
		} else if resolved {
			resolvedPaths.Add(condAllowPath)
		}
	}

	// Each allow path must be in both sets
	for _, identityAllowPath := range *identityAllow {
		resolvedPaths.Add(identityAllowPath)
	}
	for _, resourceAllowPath := range *resourceAllow {
		resolvedPaths.Add(resourceAllowPath)
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

func GetAWSNodeByGraphID(ctx context.Context, db graph.Database, id graph.ID) (*graph.Node, error) {
	var node *graph.Node
	var err error

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		node, err = ops.FetchNode(tx, id)
		return nil
	})

	return node, err
}

func GetAWSNodesByKind(ctx context.Context, db graph.Database, nodeKind graph.Kind) (cardinality.Duplex[uint32], error) {
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
