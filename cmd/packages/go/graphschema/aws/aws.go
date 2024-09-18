// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by Cuelang code gen. DO NOT EDIT!
// Cuelang source: github.com/specterops/bloodhound/-/tree/main/packages/cue/schemas/

package aws

import (
	graph "github.com/specterops/bloodhound/dawgs/graph"
)

var (
	AWSAccount = graph.StringKind("AWSAccount")
	AWSEntity = graph.StringKind("AWSBase")
	AWSTag = graph.StringKind("AWSTag")
	AWSAction = graph.StringKind("AWSAction")
	AWSActionBlob = graph.StringKind("AWSActionBlob")
	AWSResourceBlob = graph.StringKind("AWSResourceBlob")
	AWSConditionKey = graph.StringKind("AWSConditionKey")
	AWSConditionValue = graph.StringKind("AWSConditionValue")
	AWSConditionOperator = graph.StringKind("AWSConditionOperator")
	AWSCondition = graph.StringKind("AWSCondition")
	AWSStatement = graph.StringKind("AWSStatement")
	AWSPolicyDocument = graph.StringKind("AWSPolicyDocument")
	AWSPolicyVersion = graph.StringKind("AWSPolicyVersion")
	AWSManagedPolicy = graph.StringKind("AWSManagedPolicy")
	AWSInlinePolicy = graph.StringKind("AWSInlinePolicy")
	AWSAssumeRolePolicy = graph.StringKind("AWSAssumeRolePolicy")
	AWSRole = graph.StringKind("AWSRole")
	AWSUser = graph.StringKind("AWSUser")
	AWSGroup = graph.StringKind("AWSGroup")
	UniqueArn = graph.StringKind("UniqueArn")
	AWSResourceType = graph.StringKind("AWSResourceType")
	
	ActsOn = graph.StringKind("ActsOn")
	AllowAction = graph.StringKind("Action")
	NotAction = graph.StringKind("NotAction")
	AttachedTo = graph.StringKind("AttachedTo")
	ExpandsTo = graph.StringKind("ExpandsTo")
	Resource = graph.StringKind("Resource")
	NotResource = graph.StringKind("NotResource")
	MemberOf = graph.StringKind("MemberOf")
	TypeOf = graph.StringKind("TypeOf")
	IdentityTransform = graph.StringKind("IdentityTransform")

)

type Property string

const (
	AttachmentCount			Property = "attachmentcount"
	CreateDate				Property = "createdate"
	DefaultVersionId		Property = "defaultversionid"
	IsAttachable			Property = "isattachable"
	Path 					Property = "path"
	PermissionsBoundaryUsageCount Property = "permissionsboundaryusagecount"
	PolicyId				Property = "policyid"
	PolicyName				Property = "policyname"
	RoleId					Property = "roleid"
	RoleName				Property = "rolename"
	UpdateDate				Property = "updatedate"
)

type IdentityTrasformType string

const (
	IdentityTransformAssumeRole IdentityTrasformType = "sts:assumerole"
	IdentityTransformUpdateAssumeRolePolicy IdentityTrasformType = "iam:updateassumerolepolicy"
)