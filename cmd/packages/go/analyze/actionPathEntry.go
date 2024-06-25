package analyze

import (
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type ActionPathEntry struct {
	PrincipalID       graph.ID      `json:"principal_id"`
	PrincipalArn      string        `json:"principal_arn"`
	IsPrincipalDirect bool          `json:"is_principal_direct"`
	ResourceArn       string        `json:"resource_arn"`
	Action            string        `json:"action"`
	Path              graph.Path    `json:"path"`
	Effect            string        `json:"effect"`
	Statement         *graph.Node   `json:"statement"`
	Conditions        graph.NodeSet `json:"conditions"`
}

func (a *ActionPathEntry) IsEqual(other ActionPathEntry) bool {
	return a.PrincipalArn == other.PrincipalArn && a.Action == other.Action && a.ResourceArn == other.ResourceArn
}

func (p *ActionPathEntry) String() string {
	return fmt.Sprintf("PrincipalArn: %s, Action: %s, Effect: %s", p.PrincipalArn, p.Action, p.Effect)
}
