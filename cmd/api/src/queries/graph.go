package queries

import (
	"context"
	"net/url"
	"strings"

	"github.com/hotnops/apeman/analyze"
	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
)

type PermissionMapping struct {
	Arn     string                `json:"arn"`
	Actions map[string][]graph.ID `json:"actions"`
}

func CypherQuery(ctx context.Context, db graph.Database, cypherQuery string) (graph.PathSet, error) {

	var returnPathSet graph.PathSet

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if pathSet, err := ops.FetchPathSetByQuery(tx, cypherQuery); err != nil {
			return err
		} else {
			returnPathSet = pathSet
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return returnPathSet, nil
}

func Search(ctx context.Context, db graph.Database, searchString string) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {

			return query.Or(
				query.CaseInsensitiveStringContains(query.NodeProperty("arn"), searchString),
				query.CaseInsensitiveStringContains(query.NodeProperty("name"), searchString),
				query.CaseInsensitiveStringContains(query.NodeProperty("hash"), searchString),
			)
		})); err != nil {
			return err
		} else {
			for _, fetchedNode := range fetchedNodes {
				nodes.Add(fetchedNode)
			}
			return nil
		}
	})

	return nodes, nil
}

func GetAllAWSNodes(ctx context.Context, db graph.Database, parameters url.Values) ([]*graph.Node, error) {
	var nodes []*graph.Node

	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			criteria := make([]graph.Criteria, 0)
			for key, _ := range parameters {
				if key == "kind" {
					criteria = append(criteria, query.Kind(query.Node(), graph.StringKind(parameters.Get(key))))
				} else {
					criteria = append(criteria, query.Equals(query.NodeProperty(key), parameters.Get(key)))
				}
			}
			return query.And(criteria...)
		})); err != nil {
			return err
		} else {
			for _, fetchedNode := range fetchedNodes {
				nodes = append(nodes, fetchedNode)
			}
			return nil
		}
	})

	return nodes, nil
}

func GetAWSNodeTags(ctx context.Context, db graph.Database, nodeID graph.ID) (graph.NodeSet, error) {
	graphSet := graph.NewNodeSet()

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), aws.AWSTag),
				query.Equals(query.EndID(), nodeID),
			)
		})); err != nil {
			return err
		} else {
			graphSet.AddSet(fetchedNodes)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return graphSet, nil
}

func GetActiveAWSConditionKeys(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	// Active condition keys are the ones currently being used which means they are
	// attached to a condition
	var (
		nodes = graph.NewNodeSet()
	)

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if rels, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), aws.AWSConditionKey),
				query.Kind(query.Relationship(), aws.AttachedTo),
				query.Kind(query.End(), aws.AWSCondition),
			)
		})); err != nil {
			return err
		} else {
			for _, rel := range rels {
				node, err := ops.FetchNode(tx, rel.StartID)
				if err != nil {
					return err
				}
				nodes.Add(node)
			}
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func GetAWSResourceInboundPermissions(ctx context.Context, db graph.Database, arn string) ([]PermissionMapping, error) {
	principalToStatementMap, err := analyze.GetAWSResourceInboundPermissions(ctx, db, arn)
	if err != nil {
		return nil, err
	}
	principalToActionMap := make([]PermissionMapping, 0)

	// We have all the statments that can effectively act on the
	// principal, but now we need to filter out conditions and deny statements
	for id, statements := range principalToStatementMap {
		if node, err := GetAWSNodeByGraphID(ctx, db, id); err != nil {
			return nil, err
		} else {
			permissionMapping := new(PermissionMapping)
			permissionMapping.Arn, _ = node.Properties.Get("arn").String()
			allowActions := make(map[string][]graph.ID)
			denyActions := make(map[string][]graph.ID)
			for _, statement := range statements {
				if statement.Effect == "Allow" {
					for _, action := range statement.Actions {
						if statementList, ok := allowActions[action]; !ok {
							allowActions[action] = []graph.ID{statement.StatementID}
						} else {
							allowActions[action] = append(statementList, statement.StatementID)
						}
					}
				}
				if statement.Effect == "Deny" {
					for _, action := range statement.Actions {
						if statementList, ok := allowActions[action]; !ok {
							denyActions[action] = []graph.ID{statement.StatementID}
						} else {
							denyActions[action] = append(statementList, statement.StatementID)
						}
					}
				}

			}
			for denyAction := range denyActions {
				delete(allowActions, denyAction)
			}
			permissionMapping.Actions = allowActions
			if len(allowActions) > 0 {
				principalToActionMap = append(principalToActionMap, *permissionMapping)
			}
		}
	}

	return principalToActionMap, nil
}

func GetAWSResourceInboundPrincipalsWithAction(ctx context.Context, db graph.Database, targetArn string, action string) (map[graph.ID][]analyze.AWSStatementActions, error) {
	return analyze.GetAWSResourceInboundPrincipalsWithAction(ctx, db, targetArn, action)
}

func GetAWSNodeByKindID(ctx context.Context, db graph.Database, propertyName string, id string, kind graph.Kind) (*graph.Node, error) {
	var nodes = graph.NewNodeSet()

	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.NodeProperty(propertyName), id),
				query.Kind(query.Node(), kind),
			)
		})); err != nil {
			return err
		} else {
			nodes.AddSet(fetchedNodes)
			return nil
		}
	})

	return nodes.Slice()[0], nil
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

func GetAWSRelationshipByGraphID(ctx context.Context, db graph.Database, id graph.ID) (*graph.Relationship, error) {
	var rel *graph.Relationship
	var err error

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		rel, err = ops.FetchRelationship(tx, id)
		return nil
	})

	return rel, err
}

func GetAWSNodeEdges(ctx context.Context, db graph.Database, id graph.ID, direction graph.Direction, queryParams url.Values) ([]*graph.Relationship, error) {

	var returnValue []*graph.Relationship

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			criteria := make([]graph.Criteria, 0)
			if direction == graph.DirectionInbound {
				criteria = append(criteria, query.Equals(query.EndID(), id))
			} else if direction == graph.DirectionOutbound {
				criteria = append(criteria, query.Equals(query.StartID(), id))
			} else {
				criteria = append(criteria, query.Or(query.Equals(query.StartID(), id), query.Equals(query.EndID(), id)))
			}
			for key, _ := range queryParams {
				if key == "relkind" {
					stringKinds := queryParams[key]
					kinds := graph.StringsToKinds(stringKinds)
					criteria = append(criteria, query.KindIn(query.Relationship(), kinds...))
				}
				if key == "kind" {
					stringKinds := queryParams[key]
					kinds := graph.StringsToKinds(stringKinds)
					if direction == graph.DirectionInbound {
						criteria = append(criteria, query.KindIn(query.Start(), kinds...))
					} else if direction == graph.DirectionOutbound {
						criteria = append(criteria, query.KindIn(query.End(), kinds...))
					}
				}
			}
			return query.And(criteria...)
		})); err != nil {
			return err
		} else {
			if nil == relationships {
				returnValue = []*graph.Relationship{}
			} else {
				returnValue = relationships
			}
			return nil
		}
	})

	return returnValue, err

}

func GetPrincipalsOfPolicy(ctx context.Context, db graph.Database, policyID string) (graph.NodeSet, error) {
	var (
		node *graph.Node
		err  error
	)

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNode, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.NodeProperty("policyid"), policyID),
			)
		})); err != nil {
			return err
		} else {
			node = fetchedNode[0]
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return analyze.GetPrincipalsOfPolicy(ctx, db, node)
}

func GetPoliciesOfEntity(ctx context.Context, db graph.Database, propertyName string, id string) (graph.NodeSet, error) {

	var node *graph.Node
	var err error

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodes(tx.Nodes().Filter(
			query.Equals(
				query.NodeProperty(propertyName),
				strings.ToUpper(id),
			))); err != nil {
			return err
		} else {
			node = fetchedNodes[0]
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	nodes, err := analyze.GetAttachedKinds(ctx, db, node, graph.Kinds{aws.AWSManagedPolicy, aws.AWSInlinePolicy})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
