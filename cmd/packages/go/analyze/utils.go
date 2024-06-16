package analyze

import (
	"fmt"

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

func ActionPathSetToMap(actionSet *ActionPathSet) PrincipalToActionMap {
	actionMap := make(PrincipalToActionMap)
	for _, actionPath := range actionSet.ActionPaths {
		principal := actionPath.PrincipalArn
		action := actionPath.Action
		var actions []string
		var ok bool
		// if the principal is not in the map, add it
		if actions, ok = actionMap[principal]; !ok {
			actions = make([]string, 0)
		}
		// add the action to the principal's list if it's not already there
		actions = addUniqueItem(actions, action)
		actionMap[principal] = actions

	}
	return actionMap
}
