package analyze

import (
	"fmt"

	"github.com/hotnops/apeman/awsconditions"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type ActionPathEntry struct {
	PrincipalID       graph.ID                     `json:"principal_id"`
	PrincipalTags     map[string]string            `json:"principal_tags"`
	PrincipalArn      string                       `json:"principal_arn"`
	IsPrincipalDirect bool                         `json:"is_principal_direct"`
	ResourceArn       string                       `json:"resource_arn"`
	ResourceID        graph.ID                     `json:"resource_id"`
	ResourceTags      map[string]string            `json:"resource_tags"`
	Action            string                       `json:"action"`
	Path              graph.Path                   `json:"path"`
	Effect            string                       `json:"effect"`
	Statement         *graph.Node                  `json:"statement"`
	Conditions        []awsconditions.AWSCondition `json:"conditions"`
}

func (a *ActionPathEntry) IsEqual(other ActionPathEntry) bool {
	return a.PrincipalArn == other.PrincipalArn && a.Action == other.Action && a.ResourceArn == other.ResourceArn
}

func (p *ActionPathEntry) String() string {
	return fmt.Sprintf("PrincipalArn: %s, Action: %s, Effect: %s", p.PrincipalArn, p.Action, p.Effect)
}
