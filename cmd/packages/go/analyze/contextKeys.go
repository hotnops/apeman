package analyze

import (
	"fmt"
	"strings"
)

func PrincipalArn(entry ActionPathEntry, policyVariable string) (string, error) {
	return entry.PrincipalArn, nil
}

func PrincipalAccount(entry ActionPathEntry, policyVariable string) (string, error) {
	return GetAccountIDFromArn(entry.PrincipalArn), nil
}

func PrincipalTag(entry ActionPathEntry, policyVariable string) (string, error) {
	parts := strings.Split(policyVariable, "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid policy variable")
	}

	tagName := parts[1]

	return entry.PrincipalTags[tagName], nil
}

func NotImplemented(entry ActionPathEntry, policyVariable string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func ResourceAccount(entry ActionPathEntry, policyVariable string) (string, error) {
	return GetAccountIDFromArn(entry.ResourceArn), nil
}

func ResourceTag(entry ActionPathEntry, policyVariable string) (string, error) {
	parts := strings.Split(policyVariable, "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid policy variable")
	}

	tagName := parts[1]

	return entry.ResourceTags[tagName], nil
}

var ContextKeyFunctionMap = map[string]func(ActionPathEntry, string) (string, error){
	"aws:PrincipalArn":              PrincipalArn,
	"aws:PrincipalAccount":          PrincipalAccount,
	"aws:PrincipalOrgPaths":         NotImplemented,
	"aws:PrincipalOrgID":            NotImplemented,
	"aws:PrincipalTag":              PrincipalTag,
	"aws:PrincipalIsAWSSerivce":     NotImplemented,
	"aws:PrincipalServiceName":      NotImplemented,
	"aws:PrincipalServiceNamesList": NotImplemented,
	"aws:PrincipalType":             NotImplemented,
	"aws:userid":                    NotImplemented,
	"aws:username":                  NotImplemented,
	"aws:FederatedProvider":         NotImplemented,
	"aws:ResourceAccount":           ResourceAccount,
	"aws:ResourceOrgPaths":          NotImplemented,
	"aws:ResourceOrgID":             NotImplemented,
	"aws:ResourceTag":               ResourceTag,
}
