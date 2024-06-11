package awsconditions

import (
	"errors"

	"github.com/specterops/bloodhound/dawgs/graph"
)

func PrincipalArn(path graph.Path) (string, error) {
	prinNode := path.Nodes[0]
	if arn, err := prinNode.Properties.Get("arn").String(); err != nil {
		return "", err
	} else {
		return arn, nil

	}
}

func PrincipalAccount(path graph.Path) (string, error) {
	prinNode := path.Nodes[0]
	if account, err := prinNode.Properties.Get("account_id").String(); err != nil {
		return "", err
	} else {
		return account, nil
	}
}

func PrincipalOrgPath(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalOrgID(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalTag(path graph.Path, tagKey string) (string, error) {
	return "", errors.New("Not implemented")

}

func PrincipalIsAWSService(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalServiceName(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalServiceNamesList(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalType(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalUserID(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

func PrincipalUserName(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:FederatedProvider
func FederatedProvider(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:TokenIssueTime
func TokenIssueTime(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:MultiFactorAuthAge
func MultiFactorAuthAge(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:MultiFactorAuthPresent
func MultiFactorAuthPresent(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:Ec2InstanceSourceVpc
func Ec2InstanceSourceVpc(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:Ec2InstanceSourcePrivateIPv4
func Ec2InstanceSourcePrivateIPv4(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:SourceIdentity
func SourceIdentity(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// ec2:RoleDelivery
func Ec2RoleDelivery(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// ec2:SourceInstanceArn
func Ec2SourceInstanceArn(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// glue:RoleAssumedBy
func GlueRoleAssumedBy(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// glue:CredentialIssuingService
func GlueCredentialIssuingService(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// lambda:SourceFunctionArn
func LambdaSourceFunctionArn(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// ssm:SourceInstanceArn
func SsmSourceInstanceArn(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// identitystore:UserId
func IdentityStoreUserId(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:SourceIp
func SourceIp(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:SourceVpc
func SourceVpc(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:SourceVpce
func SourceVpce(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:VpcSourceIp
func VpcSourceIp(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:ResourceAccount
func ResourceAccount(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:ResourceOrgPaths
func ResourceOrgPaths(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:ResourceOrgID
func ResourceOrgID(path graph.Path) (string, error) {
	return "", errors.New("Not implemented")
}

// aws:ResourceTag/tag-key
func ResourceTag(path graph.Path, tagKey string) (string, error) {
	return "", errors.New("Not implemented")
}
