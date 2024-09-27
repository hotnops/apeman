package queries

import (
	"context"
	"net/url"

	"github.com/hotnops/apeman/graphschema/aws"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
)

func GetAllAWSNodes(ctx context.Context, db graph.Database, parameters url.Values) ([]*graph.Node, error) {
	var nodes []*graph.Node

	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			criteria := make([]graph.Criteria, 0)
			for key := range parameters {
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

func GetAWSNodeByGraphID(ctx context.Context, db graph.Database, id graph.ID) (*graph.Node, error) {
	var node *graph.Node
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNode, err := ops.FetchNode(tx, id); err != nil {
			return err
		} else {
			node = fetchedNode
			return nil
		}
	})
	return node, err
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
			for key := range queryParams {
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
