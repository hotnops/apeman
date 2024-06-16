package queries

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

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

	if returnPathSet == nil {
		returnPathSet = graph.NewPathSet()
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

func MergePaths(firstPath graph.Path, secondPath graph.Path) (graph.Path, error) {
	mergedPath := graph.Path{}
	mergedPath.Nodes = append(firstPath.Nodes, secondPath.Nodes...)
	mergedPath.Edges = append(firstPath.Edges, secondPath.Edges...)
	return mergedPath, nil
}

func GetAWSRoleInboundRoleAssumptionPaths(ctx context.Context, db graph.Database, roleId string) (*analyze.ActionPathSet, error) {
	query := "MATCH p=(a:AWSRole) <- [:AttachedTo] - (:AWSAssumeRolePolicy) <- [:AttachedTo] - (s:AWSStatement) - [r:Resource|ExpandsTo*1..2] -> (:AWSRole|AWSUser) WHERE a.roleid = '%s' " +
		"RETURN p"

	formatted_query := fmt.Sprintf(query, roleId)

	resourcePaths := graph.NewPathSet()

	targetNode, err := GetAWSNodeByKindID(ctx, db, "roleid", roleId, aws.AWSRole)
	if err != nil {
		return nil, err
	}

	arn, _ := targetNode.Properties.Get("arn").String()
	pathSet, _ := CypherQuery(ctx, db, formatted_query)

	statements := pathSet.AllNodes().ContainingNodeKinds(aws.AWSStatement)
	for _, statement := range statements {

		// Get the actions for the statement
		action_query := "MATCH p=(s:AWSStatement) - [:Action|ExpandsTo*1..2] -> (:AWSAction) WHERE ID(s) = %d RETURN p"
		formatted_query := fmt.Sprintf(action_query, statement.ID)
		actionPaths, _ := CypherQuery(ctx, db, formatted_query)

		condition_query := "MATCH p=(s:AWSStatement) <- [:AttachedTo] - (c:AWSCondition) WHERE ID(s) = %d RETURN p"
		formatted_query = fmt.Sprintf(condition_query, statement.ID)
		conditionPaths, err := CypherQuery(ctx, db, formatted_query)
		if err != nil {
			return nil, err
		}

		pathsWithStatement := GetPathsWithStatement(pathSet, statement)

		enrichedPath := graph.Path{}
		for _, actionPath := range actionPaths {
			enrichedPath, _ = MergePaths(actionPath, enrichedPath)
		}
		for _, conditionPath := range conditionPaths {
			enrichedPath, _ = MergePaths(conditionPath, enrichedPath)
		}

		for _, path := range pathsWithStatement {
			fullPath, _ := MergePaths(path, enrichedPath)
			resourcePaths.AddPath(fullPath)
		}
	}

	resourceActionPathSet := analyze.ResourcePolicyPathToActionPathSet(resourcePaths)
	principalArns := resourceActionPathSet.GetPrincipals()

	identityPaths := graph.NewPathSet()
	uniqueActions := resourcePaths.AllNodes().ContainingNodeKinds(aws.AWSAction)
	if len(principalArns) > 0 {
		for _, action := range uniqueActions {
			// For each action, get the identity policy permissions
			// and then get the union of these paths and the resolved paths
			actionName, err := action.Properties.Get("name").String()
			if err != nil {
				log.Printf("[!] Error getting action name: %s", err.Error())
				continue
			}
			actionPaths, err := GetIdentityPolicyPathsOnArnWithAction(ctx, db, arn, actionName, principalArns)
			if err != nil {
				log.Printf("[!] Error getting identity policy permissions: %s", err.Error())
				continue
			}
			identityPaths.AddPathSet(actionPaths)
		}
	}

	identityActionPathSet := analyze.IdentityPolicyPathToActionPathSet(identityPaths)

	// Get the intersection of the two path sets. Unlike most resource policies,
	// which are unions, assume role policy must be an intersection.
	allowedPaths, err := analyze.ActionPathSetIntersection(resourceActionPathSet, identityActionPathSet)
	if err != nil {
		return nil, err
	}

	// Now filter out identity based paths
	// For each path, check the identity policy allows the action

	// Get the intersection of the resolved paths and the identity paths
	return allowedPaths, nil
}

func CreateAssumeRoleEdge(ctx context.Context, db graph.Database, sourceNodes []graph.ID, targetNode graph.ID) error {

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, sourceNode := range sourceNodes {
			properties := graph.NewProperties()
			properties.Set("layer", 3)
			properties.Set("name", "sts:AssumeRole")
			_, err := tx.CreateRelationshipByIDs(sourceNode, targetNode, aws.IdentityTransform, properties)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func CreateAssumeRoleEdgesToRole(ctx context.Context, db graph.Database, roleNode *graph.Node, counter *Counter) {
	roleId, _ := roleNode.Properties.Get("roleid").String()
	rolePaths, err := GetAWSRoleInboundRoleAssumptionPaths(ctx, db, roleId)
	if err != nil {
		log.Printf("[!] Error getting role assumption paths: %s", err.Error())
	}

	if len(rolePaths.ActionPaths) > 0 {
		sourceIDs := make([]graph.ID, 0)
		for _, actionPath := range rolePaths.ActionPaths {
			sourceIDs = append(sourceIDs, actionPath.PrincipalID)
		}
		err := CreateAssumeRoleEdge(ctx, db, sourceIDs, roleNode.ID)
		if err != nil {
			log.Printf("[!] Error creating assume role edge: %s", err.Error())
		}

	}

	// Update and log progress
	processedCount := counter.Increment()
	if (processedCount % 100) == 0 {
		log.Printf("Processed %d out of %d", processedCount, counter.Total)
	}

}

type Counter struct {
	mu    sync.Mutex
	count int
	Total int
}

// Increment safely increments the counter and returns the new count
func (c *Counter) Increment() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	return c.count
}

func CreateAssumeRoleEdges(ctx context.Context, db graph.Database) error {
	roleNodes := graph.NewNodeSet()
	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(query.Kind(query.Node(), aws.AWSRole))
		})); err != nil {
			return err
		} else {
			for _, fetchedNode := range fetchedNodes {
				roleNodes.Add(fetchedNode)
			}
			return nil
		}
	})

	const numWorkers = 1000
	jobs := make(chan *graph.Node, len(roleNodes.Slice()))
	var wg sync.WaitGroup

	// Initialize the counter
	counter := &Counter{Total: len(roleNodes.Slice())}

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for roleNode := range jobs {
				CreateAssumeRoleEdgesToRole(ctx, db, roleNode, counter)
			}
		}()
	}

	// Distribute jobs
	for _, roleNode := range roleNodes.Slice() {
		jobs <- roleNode
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	log.Println("All jobs processed")
	return nil
}

func GetIdentityPolicyPathsOnArnWithAction(ctx context.Context, db graph.Database, arn string, actionName string, principalArns []string) (graph.PathSet, error) {

	var formattedItems strings.Builder
	for i, item := range principalArns {
		formattedItems.WriteString(fmt.Sprintf("'%s'", item))
		if i < len(principalArns)-1 {
			formattedItems.WriteString(", ")
		}
	}

	query := "MATCH (b:UniqueArn) WHERE b.arn = '%s' " +
		"MATCH (a:AWSUser|AWSRole) WHERE a.arn IN [%s] " +
		"OPTIONAL MATCH p1=(a) <- [:AttachedTo*3..4] - (s1:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s1) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: '%s'}) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH p2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s2:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s2) - [:Action|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"WITH collect(p1) + collect(p2) AS paths, collect(s1) + collect(s2) as statements, b, a WHERE paths IS NOT NULL " +
		"UNWIND statements as s " +
		"OPTIONAL MATCH pa1=(a) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction {name: '%s'}) WHERE (act) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH pa2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction {name: '%s'}) WHERE (act) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"WITH collect(pa1) + collect(pa2) AS prinToAction WHERE prinToAction IS NOT NULL " +
		"UNWIND(prinToAction) as p " +
		"RETURN p"

	formatted_query := fmt.Sprintf(query, arn, formattedItems.String(), actionName, actionName, actionName)

	pathSet, err := CypherQuery(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}

	return pathSet, nil
}

func GetResourcePathFromStatementToArn(ctx context.Context, db graph.Database, statementID graph.ID, arn string) (graph.PathSet, error) {
	query := "MATCH rp=(s:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b:UniqueArn) WHERE ID(s) = %d AND b.arn = '%s' RETURN rp"
	formatted_query := fmt.Sprintf(query, statementID, arn)

	pathSet, err := CypherQuery(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}
	return pathSet, nil
}

func GetConditionPathFromStatement(ctx context.Context, db graph.Database, statementID graph.ID) (graph.PathSet, error) {
	query := "MATCH cp=(s:AWSStatement) <- [:AttachedTo] - (c:AWSCondition) WHERE ID(s) = %d RETURN cp"
	formatted_query := fmt.Sprintf(query, statementID)

	pathSet, err := CypherQuery(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}
	return pathSet, nil
}

func GetPathsWithStatement(paths graph.PathSet, statement *graph.Node) graph.PathSet {
	pathsWithStatement := graph.NewPathSet()
	for _, path := range paths {
		if path.ContainsNode(statement.ID) {
			pathsWithStatement.AddPath(path)
		}
	}
	return pathsWithStatement
}

func GetAllIdentityPolicyPathsOnArn(ctx context.Context, db graph.Database, arn string) (*analyze.ActionPathSet, error) {

	query := "MATCH (b:UniqueArn) WHERE b.arn = '%s' " +
		"MATCH (a:AWSUser|AWSRole) WHERE a.account_id = b.account_id " +
		"OPTIONAL MATCH p1=(a) <- [:AttachedTo*3..4] - (s1:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s1) - [:Action|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH p2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s2:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s2) - [:Action|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"WITH collect(p1) + collect(p2) AS paths, collect(s1) + collect(s2) as statements, b, a WHERE paths IS NOT NULL " +
		"UNWIND statements as s " +
		"OPTIONAL MATCH pa1=(a) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction) WHERE (act) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH pa2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction) WHERE (act) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"WITH collect(pa1) + collect(pa2) AS prinToAction WHERE prinToAction IS NOT NULL " +
		"UNWIND(prinToAction) as p " +
		"RETURN p"

	formatted_query := fmt.Sprintf(query, arn)

	pathSet, err := CypherQuery(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}

	fullPaths := graph.NewPathSet()

	statements := pathSet.AllNodes().ContainingNodeKinds(aws.AWSStatement)
	for _, statement := range statements {
		// Get the resource path
		resourcePaths, err := GetResourcePathFromStatementToArn(ctx, db, statement.ID, arn)
		if err != nil {
			log.Printf("[!] Error getting resource path: %s", err.Error())
			continue
		}
		conditionPaths, err := GetConditionPathFromStatement(ctx, db, statement.ID)
		if err != nil {
			log.Printf("[!] Error getting condition path: %s", err.Error())
			continue
		}

		pathsWithStatement := GetPathsWithStatement(pathSet, statement)

		enrichedPath := graph.Path{}
		for _, resourcePath := range resourcePaths {
			enrichedPath, _ = MergePaths(resourcePath, enrichedPath)
		}
		for _, conditionPath := range conditionPaths {
			enrichedPath, _ = MergePaths(conditionPath, enrichedPath)
		}

		for _, path := range pathsWithStatement {
			fullPath, _ := MergePaths(path, enrichedPath)
			fullPaths.AddPath(fullPath)
		}

	}

	resolvedPaths, err := analyze.ResolvePaths(fullPaths)
	if err != nil {
		return nil, err
	}

	return analyze.IdentityPolicyPathToActionPathSet(resolvedPaths), nil

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
