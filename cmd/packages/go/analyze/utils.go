package analyze

import (
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
)

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

func ResolveConditons(path graph.Path) (bool, error) {
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
