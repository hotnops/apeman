package queries

import (
	"context"
	"fmt"
	"log"
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

func RawCypherQuery(ctx context.Context, db graph.Database, query string, paramaters map[string]any) ([]graph.ValueMapper, error) {
	var values []graph.ValueMapper
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if result := tx.Run(query, paramaters); result.Error() != nil {
			return result.Error()
		} else {
			for result.Next() {
				nextValues := result.Values()
				values = append(values, nextValues)
			}
			return nil
		}
	}); err != nil {
		return nil, err
	}
	return values, nil
}

func CypherQueryPaths(ctx context.Context, db graph.Database, cypherQuery string) (graph.PathSet, error) {

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

func GetAWSAccountIDs(ctx context.Context, db graph.Database) ([]string, error) {
	var accountIDs []string
	query := "MATCH (a:UniqueArn) RETURN DISTINCT a.account_id"
	results, err := RawCypherQuery(ctx, db, query, nil)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		var accountID string
		err = result.Map(&accountID)
		if err != nil {
			continue
		}
		if accountID == "" {
			continue
		}
		accountIDs = append(accountIDs, accountID)
	}
	return accountIDs, nil
}

func GetAWSAccountServices(ctx context.Context, db graph.Database, accountID string) ([]string, error) {
	var services []string

	query := "MATCH (a:UniqueArn) WHERE a.account_id = $account_id RETURN DISTINCT a.service"
	params := map[string]interface{}{"account_id": accountID}

	results, err := RawCypherQuery(ctx, db, query, params)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		var service string
		err = result.Map(&service)
		if err != nil {
			continue
		}
		services = append(services, service)
	}

	return services, nil

}

func GetAWSRoleInboundRoleAssumptionPaths(ctx context.Context, db graph.Database, roleId string) (*analyze.ActionPathSet, error) {
	// First, get all the principals that are trusted to assume this role
	query := "MATCH p=(a:AWSRole) <- [:AttachedTo] - (:AWSAssumeRolePolicy) <- [:AttachedTo] - (s:AWSStatement) - [:Principal|ExpandsTo*1..2] -> (b:AWSRole|AWSUser) WHERE a.roleid = $roleid AND (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name:'sts:assumerole'}) " +
		"WITH a, s, b " +
		"OPTIONAL MATCH (s) <- [:AttachedTo] - (c:AWSCondition) " +
		"OPTIONAL MATCH (s) - [:Principal] - > (pb:AWSPrincipalBlob) - [:ExpandsTo*1..2] -> (b) " +
		"RETURN b, a.arn, s, COALESCE(c IS NOT NULL, false), COALESCE(pb IS NOT NULL, false)"

	params := map[string]any{
		"roleid": roleId,
	}

	results, err := RawCypherQuery(ctx, db, query, params)
	if err != nil {
		return nil, err
	}

	resourcePathSet := analyze.ActionPathSet{}

	for _, result := range results {
		newActionPathEntry := analyze.ActionPathEntry{}
		var sourceNode graph.Node
		var destArn string
		var statement graph.Node
		var conditionExists bool
		var isPrinExpanded bool

		err = result.Map(&sourceNode)
		if err != nil {
			continue
		}
		err = result.Map(&destArn)
		if err != nil {
			continue
		}
		err = result.Map(&statement)
		if err != nil {
			continue
		}
		err = result.Map(&conditionExists)
		if err != nil {
			continue
		}

		err = result.Map(&isPrinExpanded)
		if err != nil {
			continue
		}

		effect, _ := statement.Properties.Get("effect").String()

		if conditionExists {
			conditions, err := GetConditionPathsFromStatement(ctx, db, statement.ID)
			if err != nil {
				log.Printf("[!] Error getting conditions: %s", err.Error())
				continue
			}
			newActionPathEntry.Conditions = conditions
		}
		newActionPathEntry.PrincipalID = sourceNode.ID
		sourceArn, _ := sourceNode.Properties.Get("arn").String()
		newActionPathEntry.PrincipalArn = sourceArn
		newActionPathEntry.ResourceArn = destArn
		newActionPathEntry.Effect = effect
		newActionPathEntry.Statement = &statement
		newActionPathEntry.Action = "sts:assumerole"
		newActionPathEntry.IsPrincipalDirect = !isPrinExpanded
		resourcePathSet.Add(newActionPathEntry)
	}

	// Now, we need all the identity paths from these principals to the role
	principals := []string{}

	for _, entry := range resourcePathSet {
		principals = append(principals, entry.PrincipalArn)
	}

	if len(principals) == 0 {
		return nil, nil
	}

	identityPaths, err := GetAllUnresolvedIdentityPolicyPathsOnArnWithArnsAndActions(ctx, db, roleId, "sts:assumerole", principals)
	if err != nil {
		return nil, err
	}

	resolvedPaths, err := analyze.ResolveAssumeRolePaths(&resourcePathSet, identityPaths)
	if err != nil {
		return nil, err
	}

	return resolvedPaths, nil
}

func CreateIdentityTransformEdge(ctx context.Context, db graph.Database, sourceNodes []graph.ID, targetNode graph.ID, name string) error {
	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, sourceNode := range sourceNodes {
			properties := graph.NewProperties()
			properties.Set("layer", 2)
			properties.Set("name", name)
			_, err := tx.CreateRelationshipByIDs(sourceNode, targetNode, aws.IdentityTransform, properties)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func CreateAssumeRoleEdgesToRole(ctx context.Context, db graph.Database, roleNode *graph.Node, counter *Counter) {
	roleId, _ := roleNode.Properties.Get(string(aws.RoleId)).String()
	roleArn, _ := roleNode.Properties.Get("arn").String()
	rolePaths, err := GetAWSRoleInboundRoleAssumptionPaths(ctx, db, roleId)
	if err != nil {
		log.Printf("[!] Error getting role assumption paths: %s", err.Error())
	}
	if rolePaths == nil {
		log.Printf("[!] No role assumption paths found for role %s", roleArn)
		return
	}

	if len(*rolePaths) > 0 {
		sourceIDs := make([]graph.ID, 0)
		for _, actionPath := range *rolePaths {
			log.Printf("[*] Creating assume role edge from %s to %s", actionPath.PrincipalArn, actionPath.ResourceArn)
			sourceIDs = append(sourceIDs, actionPath.PrincipalID)
		}

		err := CreateIdentityTransformEdge(ctx, db, sourceIDs, roleNode.ID, string(aws.IdentityTransformAssumeRole))
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
	log.Printf("[*] Getting all nodes")
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

	log.Printf("[*] Found %d roles", len(roleNodes.Slice()))

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

func GetUnresolvedOutputPaths(ctx context.Context, db graph.Database, principalNode *graph.Node) (*analyze.ActionPathSet, error) {
	// First, get all resources that this principal has a path to, regardless of deny or allow
	query := "MATCH p=(a:AWSUser|AWSRole) <- [:AttachedTo] - (:AWSManagedPolicy|AWSInlinePolicy) <- [:AttachedTo*2..3] - (s:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b:UniqueArn) " +
		"WHERE ID(a) = %d AND a.account_id = b.account_id  " +
		"WITH a, s, b " +
		"MATCH p2=(a) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction) " +
		"WHERE (act) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH (s) <- [:AttachedTo] - (c:AWSCondition) " +
		"RETURN a.arn, b.arn, s, act.name, COALESCE(c IS NOT NULL, false)"

	formatted_query := fmt.Sprintf(query, principalNode.ID)
	actionPathSet := analyze.ActionPathSet{}
	result, err := RawCypherQuery(ctx, db, formatted_query, nil)
	for _, item := range result {
		var sourcearn string
		var destArn string
		var statement graph.Node
		var action string
		var conditionExists bool
		err = item.Map(&sourcearn)
		if err != nil {
			continue
		}
		err = item.Map(&destArn)
		if err != nil {
			continue
		}
		err = item.Map(&statement)
		if err != nil {
			continue
		}
		err = item.Map(&action)
		if err != nil {
			continue
		}
		err = item.Map(&conditionExists)
		if err != nil {
			continue
		}

		effect, _ := statement.Properties.Get("effect").String()

		entry := analyze.ActionPathEntry{}
		entry.PrincipalID = principalNode.ID
		entry.PrincipalArn = sourcearn
		entry.ResourceArn = destArn
		entry.Action = action
		entry.Effect = effect
		if conditionExists {
			conditions, err := GetConditionPathsFromStatement(ctx, db, statement.ID)
			if err != nil {
				log.Printf("[!] Error getting conditions: %s", err.Error())
				continue
			}
			entry.Conditions = conditions
		}

		entry.Conditions = nil
		actionPathSet.Add(entry)
	}
	if err != nil {
		log.Printf("[!] Error getting paths: %s", err.Error())
		return nil, err
	}

	return &actionPathSet, nil

}

func GetActionPathsFromStatementToArn(ctx context.Context, db graph.Database, statementID graph.ID, arn string) (graph.PathSet, error) {
	query := "MATCH ap=(s:AWSStatement) - [:Action|ExpandsTo*1..2] -> (a:AWSAction) WHERE ID(s) = %d AND (a) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (:UniqueArn {arn: '%s'}) RETURN ap"
	formatted_query := fmt.Sprintf(query, statementID, arn)

	pathSet, err := CypherQueryPaths(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}
	return pathSet, nil
}

func GetResourcePathFromStatementToArn(ctx context.Context, db graph.Database, statementID graph.ID, arn string) (graph.PathSet, error) {
	query := "MATCH rp=(s:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b:UniqueArn) WHERE ID(s) = %d AND b.arn = '%s' RETURN rp"
	formatted_query := fmt.Sprintf(query, statementID, arn)

	pathSet, err := CypherQueryPaths(ctx, db, formatted_query)
	if err != nil {
		return nil, err
	}
	return pathSet, nil
}

func GetConditionPathsFromStatement(ctx context.Context, db graph.Database, statementID graph.ID) (graph.NodeSet, error) {
	query := "MATCH (s:AWSStatement) <- [:AttachedTo] - (c:AWSCondition) WHERE ID(s) = %d RETURN c"
	formatted_query := fmt.Sprintf(query, statementID)

	results, err := RawCypherQuery(ctx, db, formatted_query, nil)
	if err != nil {
		return nil, err
	}

	conditions := graph.NewNodeSet()

	for _, result := range results {
		var condition graph.Node
		err = result.Map(&condition)
		if err != nil {
			continue
		}
		conditions.Add(&condition)
	}

	return conditions, nil
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

func GetAllUnresolvedIdentityPolicyPathsOnArnWithArnsAndActions(ctx context.Context, db graph.Database, roleId string, actionName string, sourceArns []string) (*analyze.ActionPathSet, error) {

	query := "MATCH (b:AWSRole) " +
		"WHERE b.roleid = $roleId " +
		"MATCH (a:AWSUser|AWSRole) " +
		"WHERE a.arn in $sourceArns " +
		"MATCH p1=(a) <- [:AttachedTo*3..4] - (s:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) " +
		"WITH a, s, b " +
		"MATCH p2 = (s) - [:Action|ExpandsTo*1..2] -> (act:AWSAction {name: $actionName}) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH (s) <- [:AttachedTo] - (c:AWSCondition) " +
		"RETURN a, b.arn, s, act.name, COALESCE(c IS NOT NULL, false)"

	params := map[string]any{
		"roleId":     roleId,
		"sourceArns": sourceArns,
		"actionName": actionName,
	}

	results, err := RawCypherQuery(ctx, db, query, params)
	if err != nil {
		return nil, err
	}
	actionPathSet := analyze.ActionPathSet{}
	for _, result := range results {
		newActionPathEntry := analyze.ActionPathEntry{}
		var sourceNode graph.Node
		var destArn string
		var statement graph.Node
		var action string
		var conditionExists bool
		err = result.Map(&sourceNode)
		if err != nil {
			continue
		}
		err = result.Map(&destArn)
		if err != nil {
			continue
		}
		err = result.Map(&statement)
		if err != nil {
			continue
		}
		err = result.Map(&action)
		if err != nil {
			continue
		}
		err = result.Map(&conditionExists)
		if err != nil {
			continue
		}

		effect, _ := statement.Properties.Get("effect").String()

		if conditionExists {
			conditions, err := GetConditionPathsFromStatement(ctx, db, statement.ID)
			if err != nil {
				log.Printf("[!] Error getting conditions: %s", err.Error())
				continue
			}
			newActionPathEntry.Conditions = conditions
		}

		newActionPathEntry.PrincipalID = sourceNode.ID
		sourceArn, _ := sourceNode.Properties.Get("arn").String()
		newActionPathEntry.PrincipalArn = sourceArn
		newActionPathEntry.ResourceArn = destArn
		newActionPathEntry.Effect = effect
		newActionPathEntry.Action = action
		actionPathSet.Add(newActionPathEntry)
	}

	return &actionPathSet, nil

}

func GetAllUnresolvedIdentityPolicyPathsOnArn(ctx context.Context, db graph.Database, arn string) (*analyze.ActionPathSet, error) {

	query := "MATCH (b:UniqueArn) WHERE b.arn = '%s' " +
		"MATCH (a:AWSUser|AWSRole) WHERE a.account_id = b.account_id " +
		"OPTIONAL MATCH p1=(a) <- [:AttachedTo*3..4] - (s1:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s1) - [:Action|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH p2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s2:AWSStatement) - [:Resource|ExpandsTo*1..2] -> (b) WHERE (s2) - [:Action|ExpandsTo*1..2] -> (:AWSAction) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"WITH collect(p1) + collect(p2) AS paths, collect(s1) + collect(s2) as statements, b, a WHERE paths IS NOT NULL " +
		"UNWIND statements as s " +
		"OPTIONAL MATCH pa1=(a) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act1:AWSAction) WHERE (act1) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH pa2=(a) - [:MemberOf] -> (:AWSGroup) <- [:AttachedTo*3..4] - (s) - [:Action|ExpandsTo*1..2] -> (act2:AWSAction) WHERE (act2) - [:ActsOn] -> (:AWSResourceType) <- [:TypeOf] - (b) " +
		"OPTIONAL MATCH (s) <- [:AttachedTo] - (c:AWSCondition) " +
		"RETURN a.arn, b.arn, s, COALESCE(act1.name, act2.name), COALESCE(c IS NOT NULL, false)"

	formatted_query := fmt.Sprintf(query, arn)

	log.Printf("%s", formatted_query)
	results, err := RawCypherQuery(ctx, db, formatted_query, nil)
	if err != nil {
		return nil, err
	}
	actionPathSet := analyze.ActionPathSet{}
	for _, result := range results {
		newActionPathEntry := analyze.ActionPathEntry{}
		var sourceArn string
		var destArn string
		var statement graph.Node
		var action string
		var conditionExists bool
		err = result.Map(&sourceArn)
		if err != nil {
			continue
		}
		err = result.Map(&destArn)
		if err != nil {
			continue
		}
		err = result.Map(&statement)
		if err != nil {
			continue
		}
		err = result.Map(&action)
		if err != nil {
			continue
		}
		err = result.Map(&conditionExists)
		if err != nil {
			continue
		}

		effect, _ := statement.Properties.Get("effect").String()

		if conditionExists {
			conditions, err := GetConditionPathsFromStatement(ctx, db, statement.ID)
			if err != nil {
				log.Printf("[!] Error getting conditions: %s", err.Error())
				continue
			}
			newActionPathEntry.Conditions = conditions
		}

		newActionPathEntry.PrincipalArn = sourceArn
		newActionPathEntry.ResourceArn = destArn
		newActionPathEntry.Effect = effect
		newActionPathEntry.Action = action
		actionPathSet.Add(newActionPathEntry)
	}

	return &actionPathSet, nil

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
