package analyze

import (
	"fmt"
	"log"
	"strings"

	"github.com/hotnops/apeman/awsconditions"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type PrincipalToActionMap map[string][]string
type ActionToPathMap map[string][]ActionPathEntry

func GetNodeFromPathByKind(path graph.Path, kind graph.Kind) *graph.Node {
	for _, node := range path.Nodes {
		if node.Kinds.ContainsOneOf(kind) {
			return node
		}
	}
	return nil
}

func GetNodesFromPathByKind(path graph.Path, kind graph.Kind) graph.NodeSet {
	nodes := graph.NewNodeSet()
	for _, node := range path.Nodes {
		if node.Kinds.ContainsOneOf(kind) {
			nodes.Add(node)
		}
	}
	return nodes
}

func RemovePathByIndex(g *graph.PathSet, index int) {
	(*g) = append((*g)[:index], (*g)[index+1:]...)
}

func ResolvePolicyVariable(entry ActionPathEntry, policyVariable string) (string, error) {
	// Extract the variable between the ${} delimeter
	// Trim anything to right of / if it exists

	// If the context key has a forward slash, take the first part
	parts := strings.Split(policyVariable, "/")
	name := parts[0]

	contextResolveFunction, ok := ContextKeyFunctionMap[name]
	if !ok {
		return "", fmt.Errorf("context key %s not found", name)
	}
	return contextResolveFunction(entry, policyVariable)

}

func ResolveConditionVariables(entry ActionPathEntry, condition awsconditions.AWSCondition) (map[string][]string, error) {
	// Resolve the policy variable to the actual value
	// The policy variable is the key in the condition
	// The value is the value in the condition
	// The condition is the condition in the policy
	// The entry is the path entry

	resolvedConditionKeys := make(map[string][]string)
	var err error

	for conditionKey, conditionValues := range condition.ConditionKeys {
		// If condition key contains ${}
		conditionKey, err = ResolvePolicyVariable(entry, conditionKey)
		if err != nil {
			return nil, err
		}
		resolvedConditionValues := []string{}
		for _, conditionValue := range conditionValues {
			if strings.Contains(conditionValue, "${") {
				// Remove the value between the ${}
				trimmedValue := strings.Trim(conditionValue, "${")
				trimmedValue = strings.Trim(trimmedValue, "}")

				conditionValue, err = ResolvePolicyVariable(entry, trimmedValue)
				if err != nil {
					log.Printf("Error resolving policy variable: %s", err.Error())
					continue
				}
			}
			resolvedConditionValues = append(resolvedConditionValues, conditionValue)
		}

		resolvedConditionKeys[conditionKey] = resolvedConditionValues
	}

	return resolvedConditionKeys, nil
}

func ResolveConditions(entry ActionPathEntry) (bool, error) {
	// "The difference between single-valued and multivalued context keys depends on the number of values
	// in the request context, not the number of values in the policy condition."

	conditions := entry.Conditions
	var err error

	for _, condition := range conditions {
		// Each operator is AND'd together, so if one is fales, the condition
		// set fails
		condition.ConditionKeys, err = ResolveConditionVariables(entry, condition)
		if err != nil {
			// An error means that the variable doesn't exist or
			// can't be resolved. Simply return false, as it implies
			// a failed condition
			return false, nil
		}
		if !awsconditions.SolveCondition(&condition) {
			return false, nil
		}
	}

	return true, nil
}

func GetNodeFromPathByID(path graph.Path, id graph.ID) (*graph.Node, error) {
	for _, node := range path.Nodes {
		if node.ID == id {
			return node, nil
		}
	}
	return nil, fmt.Errorf("Node with ID %s not found in path", id)
}

func addUniqueItem(slice []string, item string) []string {
	for _, v := range slice {
		if v == item {
			// Item already exists, return the original slice
			return slice
		}
	}
	// Item does not exist, append it
	return append(slice, item)
}

func extractAccountID(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("Invalid ARN format")
	}
	accountID := parts[4]
	return accountID, nil

}
