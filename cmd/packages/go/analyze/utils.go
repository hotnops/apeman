package analyze

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type PrincipalToActionMap map[string][]string

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

func ResolveConditons(entry ActionPathEntry) (bool, error) {
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
