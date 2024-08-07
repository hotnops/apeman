package analyze

import (
	"github.com/specterops/bloodhound/dawgs/graph"
)

type ActionPathSet []ActionPathEntry

func (a *ActionPathSet) Add(actionPath ActionPathEntry) {
	*a = append(*a, actionPath)
}

func (a *ActionPathSet) AddPathSet(actionPathSet ActionPathSet) {
	for _, actionPath := range actionPathSet {
		a.Add(actionPath)
	}
}

func (a *ActionPathSet) ContainsActionPath(actionPath ActionPathEntry) bool {
	for _, path := range *a {
		if path.IsEqual(actionPath) {
			return true
		}
	}
	return false
}

func (a *ActionPathSet) GetPaths() graph.PathSet {
	paths := graph.NewPathSet()
	for _, actionPath := range *a {
		paths.AddPath(actionPath.Path)
	}
	return paths
}

func (a *ActionPathSet) GetPrincipals() []string {
	principals := []string{}

	for _, actionPath := range *a {
		principals = append(principals, actionPath.PrincipalArn)
	}

	return principals
}

func (a *ActionPathSet) RemoveActionPathEntry(actionPath ActionPathEntry) {
	tempPaths := make([]ActionPathEntry, 0)
	for _, path := range *a {
		if !path.IsEqual(actionPath) {
			tempPaths = append(tempPaths, path)
		}
	}
	*a = tempPaths
}

func (p *ActionPathSet) SplitByEffect() (allow *ActionPathSet, deny *ActionPathSet) {
	allow = new(ActionPathSet)
	deny = new(ActionPathSet)
	for _, actionPath := range *p {
		if actionPath.Effect == "Allow" {
			allow.Add(actionPath)
		} else {
			deny.Add(actionPath)
		}
	}
	return allow, deny
}

func (p *ActionPathSet) SplitByConditionalEffect() (*ActionPathSet, *ActionPathSet, *ActionPathSet, *ActionPathSet) {
	allowMap := new(ActionPathSet)
	denyMap := new(ActionPathSet)
	condtionalAllowMap := new(ActionPathSet)
	conditionalDenyMap := new(ActionPathSet)

	for _, actionPath := range *p {
		isConditional := len(actionPath.Conditions) > 0
		if actionPath.Effect == "Allow" {
			if isConditional {
				condtionalAllowMap.Add(actionPath)
			} else {
				allowMap.Add(actionPath)
			}
		} else {
			if isConditional {
				conditionalDenyMap.Add(actionPath)
			} else {
				denyMap.Add(actionPath)
			}
		}
	}
	return allowMap, denyMap, condtionalAllowMap, conditionalDenyMap
}

func GetResourceArnsFromActionSet(actionSet ActionPathSet) []string {
	principals := make([]string, 0)
	tempDict := make(map[string]bool)
	// This is a hack to keep the items unique
	for _, actionPath := range actionSet {
		tempDict[actionPath.ResourceArn] = true
	}
	for key := range tempDict {
		principals = append(principals, key)
	}
	return principals
}

func GetPrincipalNodeIDsFromActionSet(actionSet ActionPathSet) []graph.ID {
	// Get all the principal nodes from the action path set
	graphNodeIDs := make([]graph.ID, 0)

	for _, actionPathEntry := range actionSet {
		graphNodeIDs = append(graphNodeIDs, actionPathEntry.PrincipalID)
	}

	return graphNodeIDs

}

func GetActionMapFromPathSet(actionSet ActionPathSet) (ActionToPathMap, error) {
	actionMap := make(ActionToPathMap)

	for _, actionPath := range actionSet {
		action := actionPath.Action
		// if the action is not in the map, add it
		paths, ok := actionMap[action]
		if !ok {
			paths = make([]ActionPathEntry, 0)
		}

		paths = append(paths, actionPath)
		// add the principal to the action's list if it's not already there
		actionMap[action] = paths
	}

	return actionMap, nil
}

func ResourcePathSetToMap(actionSet ActionPathSet) PrincipalToActionMap {
	actionMap := make(PrincipalToActionMap)
	for _, actionPath := range actionSet {
		principal := actionPath.ResourceArn
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

func ActionPathSetToMap(actionSet ActionPathSet) PrincipalToActionMap {
	actionMap := make(PrincipalToActionMap)
	for _, actionPath := range actionSet {
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
