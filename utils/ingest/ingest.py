#!/usr/bin/env python3
import argparse
import csv
import glob
import json
import neo4j
import os
import re
import shutil
import sys
import xxhash

from neo4j import GraphDatabase

import arn

condition_map = {}
condition_key_map = {}
condition_value_map = {}
statements_map = {}
permission_documents_map = {}
managed_policy_map = {}
policy_version_map = {}
roles_map = {}
inline_policy_hash_map = {}
trust_policy_map = {}
user_map = {}
group_map = {}
action_blob_map = {}
resource_blob_map = {}
tag_map = {}
identity_provider_map = {}
principal_blob_map = {}

hash_to_hash_rels = {}
hash_to_arn_rels = {}
arn_to_arn_rels = {}
member_of_rels = {}
operator_to_condition_rels = {}
multi_operator_to_condition_rels = {}
statement_to_action_rels = {}
statement_to_not_action_rels = {}
statement_to_resource_rels = {}
statement_to_not_resource_rels = {}
statement_to_resource_blob_rels = {}
statement_to_not_resource_blob_rels = {}
statement_to_action_blob_rels = {}
statement_to_not_action_blob_rels = {}
statement_to_principal_rels = {}
statement_to_principal_blob_rels = {}
statement_to_uniquename_rels = {}
condition_key_to_resource_rels = {}
condition_value_to_key_rels = {}

def get_hash(item_to_hash: dict):
    return xxhash.xxh128_hexdigest(json.dumps(item_to_hash, sort_keys=True))


def process_condition_value(condition_value):
    if not condition_value:
        condition_value = ""
    if condition_value not in condition_value_map:
        condition_value_map[condition_value] = {'name': condition_value}


def add_to_rels(rels_dict, key, value):
    rels = rels_dict.get(key, set([]))
    rels.add(value)
    rels_dict[key] = rels


def process_tags(principal):
    arn = principal.get("Arn", None)
    tags = principal.get("Tags", [])
    if not arn:
        return
    if not tags:
        return
    for tag in tags:
        tag_hash = get_hash(tag)
        if tag_hash not in tag_map:
            tag_map[tag_hash] = {
                'hash': tag_hash,
                'key': tag['Key'],
                'value': tag['Value']
            }
        add_to_rels(hash_to_arn_rels, tag_hash, arn)


def process_condition(operator: str, condition_keyvalue: dict):
    condition_hash = get_hash({operator: condition_keyvalue})
    # {Operator: {ConditionKey: [ConditionValue]}}

    if condition_hash not in condition_map:
        condition_map[condition_hash] = {"hash": condition_hash, 'sid': condition_keyvalue.get('sid', "")}

    add_to_rels(operator_to_condition_rels, operator, condition_hash)
    
    for condition_key, condition_values in condition_keyvalue.items():
        condition_key_hash = get_hash({condition_key: condition_values})
        if condition_key_hash not in condition_key_map:
            condition_key_map[condition_key_hash] = {'hash': condition_key_hash, 'name': condition_key}
    
        add_to_rels(hash_to_hash_rels, condition_key_hash, condition_hash)

        if type(condition_values) != list:
            condition_values = [condition_values]
        for condition_value in condition_values:
            process_condition_value(condition_value)
            add_to_rels(condition_value_to_key_rels, condition_value, condition_key_hash)

    return condition_hash


def neo4j_escape_regex(unescaped_string: str):
    unescaped_string = unescaped_string.replace(".", "\\.")
    unescaped_string = unescaped_string.replace("*", ".*")
    unescaped_string = unescaped_string.replace("?", "\\?")
    unescaped_string = unescaped_string.replace("[", "\\[")
    return unescaped_string

def has_policy_variable(resource: str):
    if "${" in resource:
        return True
    
def extract_policy_variables(input_string: str):
    # Define the pattern to match the values between ${ and }
    pattern = r'\$\{(.*?)\}'
    
    # Find all occurrences of the pattern in the input string
    matches = re.findall(pattern, input_string)
    
    return matches

def replace_policy_var_with_wildcard(input_string: str):
    # Define the pattern to match the segment starting with ${ and ending with }
    pattern = r'\$\{.*?\}'
    
    # Replace the matched pattern with an asterisk
    result = re.sub(pattern, '*', input_string)
    
    return result

def isAccountNumber(value: str):
    if len(value) == 12 and value.isdigit():
        return True
    return False

def process_principals(statement_hash, principals: dict, negated: bool):
    

    awsPrins = principals.get("AWS", [])
    services = principals.get("Service", [])
    federated = principals.get("Federated", [])
    canonical_user = principals.get("CanonicalUser", [])

    if awsPrins:
        if not type(awsPrins) == list:
            awsPrins = [awsPrins]
        for principal in awsPrins:
            if principal == "*":
                if "*" not in principal_blob_map:
                    principal_blob_map["*"] = {"name": "*", "regex": neo4j_escape_regex("*")}
                add_to_rels(statement_to_principal_blob_rels, statement_hash, "*")
            if arn.Arn.is_arn(principal):
                prinArn = arn.Arn.fromstring(principal)
                if principal.endswith(":root"):
                    name = principal.replace("root", "*")
                    if name not in principal_blob_map:
                        principal_blob_map[name] = {"name": name, "regex": neo4j_escape_regex(name)}
                    add_to_rels(statement_to_principal_blob_rels, statement_hash, name)
                else:
                    if statement_hash not in statement_to_principal_rels:
                        statement_to_principal_rels[statement_hash] = set([])
                    statement_to_principal_rels[statement_hash].add(principal)

            elif isAccountNumber(principal):
                name = f"arn:aws:iam::{principal}:*"
                if name not in principal_blob_map:
                    principal_blob_map[name] = {"name": name, "regex": neo4j_escape_regex(name)}
                add_to_rels(statement_to_principal_blob_rels, statement_hash, name)
            else:
                print(f"[*] Invalid principal: {principal}")
            
    
    if services:
        if not type(services) == list:
            services = [services]
        for service in services:
            if statement_hash not in statement_to_uniquename_rels:
                statement_to_uniquename_rels[statement_hash] = set([])
            statement_to_uniquename_rels[statement_hash].add(service)

    
    if federated:
        if not type(federated) == list:
            federated = [federated]
        for federated_principal in federated:
            if federated_principal not in identity_provider_map:
                identity_provider_map[federated_principal] = {'name': federated_principal}

            if statement_hash not in statement_to_uniquename_rels:
                statement_to_uniquename_rels[statement_hash] = set([])
            statement_to_uniquename_rels[statement_hash].add(federated_principal)

    if canonical_user:
        print("[*] Canonical user not implemented")


def process_resources(statement_hash, resources: list, negated: bool):
    for resource in resources:
        regex = None
        policy_var = False        


        if arn.Arn.is_arn(resource):
            resource_arn = arn.Arn.fromstring(resource)
            policy_var = has_policy_variable(resource)

        if "*" in resource or policy_var:
            if resource not in resource_blob_map:
                if policy_var:
                    vars = extract_policy_variables(resource)
                    temp_resource = replace_policy_var_with_wildcard(resource)
                    regex = neo4j_escape_regex(temp_resource)
                    for var in vars:
                        add_to_rels(condition_key_to_resource_rels, var, resource)
                elif "*" in resource:
                    regex = neo4j_escape_regex(resource)
                resource_blob_map[resource] = {
                    'name': resource,
                    'regex': regex
                }
            if negated:
                add_to_rels(statement_to_not_resource_blob_rels, statement_hash,
                            resource)
            else:
                add_to_rels(statement_to_resource_blob_rels, statement_hash,
                        resource)
        else:
            # Some statements will have a principal ID instead of
            # an ARN in the principal map block. This is indicative
            # of a role that has been deleted.
            # See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-roles
            if arn.Arn.is_arn(resource):
                if negated:
                    add_to_rels(statement_to_not_resource_rels, statement_hash,
                                resource)
                else:
                    add_to_rels(statement_to_resource_rels, statement_hash, resource)
            else:
                print(f"[*] Invalid resource ARN: {resource}")

def process_actions(actions: list, effect: str, statement_hash: str, negated: bool):
    for action in actions:
        action = action.lower()
        if "*" in action:
            # If a specific AWS action is not defined,
            # we will create a new ActionBlob node that
            # will link to all of the actions that it
            # encompasses
            if action not in action_blob_map:
                action_blob_map[action] = {
                    "name": action,
                    "regex": neo4j_escape_regex(action)
                }
            
            if negated:
                add_to_rels(statement_to_not_action_blob_rels,
                            statement_hash,
                            action)
            else:
                add_to_rels(statement_to_action_blob_rels,
                            statement_hash,
                            action)             
        else:
            # Action directly specified in statement
            if negated:
                add_to_rels(statement_to_not_action_rels,
                            statement_hash,
                            action)
            else:
                add_to_rels(statement_to_action_rels,
                            statement_hash,
                            action)


def process_statement(statement):
    statement_hash = get_hash(statement)

    if statement_hash in statements_map:
        print("[*] Statement already exists")
        # This statement has already been processed
        return statement_hash

    effect = statement.get("Effect", "")
    statement['hash'] = statement_hash
    statements_map[statement_hash] = statement
    conditions = statement.get("Condition", None)
    if conditions:
        for operator, value in conditions.items():
            condition_hash = process_condition(operator.lower(), value)
            add_to_rels(hash_to_hash_rels, condition_hash, statement_hash)

    actions = statement.get('Action', [])
    notActions = statement.get('NotAction', [])
    if not type(actions) == list:
        actions = [actions]
    if not type(notActions) == list:
        notActions = [notActions]

    process_actions(actions, effect, statement_hash, False)
    process_actions(notActions, effect, statement_hash, True)

    resources = statement.get('Resource', [])
    if not type(resources) == list:
        resources = [resources]
    notResources = statement.get('NotResource', [])
    if not type(notResources) == list:
        notResources = [notResources]

    process_resources(statement_hash, resources, False)
    process_resources(statement_hash, notResources, True)
    
    # # This is for statements in trust policies
    principals = statement.get('Principal', {})
    # See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-anonymous
    # If the principal is *, it means that the principal is anonymous and equivalent to
    # "AWS": ""
    if principals == "*":
        principals = {"AWS": "*"}
    process_principals(statement_hash, principals, False)

    return statement_hash


def process_permission_document(policy_document):
    document_hash = xxhash.xxh128_hexdigest(json.dumps(policy_document))

    if document_hash in permission_documents_map:
        print("[*] Document already exists")
        return document_hash

    policy_document['hash'] = document_hash
    permission_documents_map[document_hash] = policy_document

    statements = policy_document["Statement"]
    if not type(statements) == list:
        statements = [statements]
    for statement in statements:
        statement_hash = process_statement(statement)
        add_to_rels(hash_to_hash_rels, statement_hash, document_hash)

    return document_hash


def process_managed_permission_policy_versions(policy_versions):
    for policy_version in policy_versions:
        if policy_version['IsDefaultVersion']:
            hash = get_hash(policy_version)
            if hash in policy_version_map:
                print("[*] PolicyVersion already exists")
                return [hash]
            policy_version['hash'] = hash
            policy_version_map[hash] = policy_version
            policy_document = policy_version["Document"]
            document_hash = process_permission_document(policy_document)
            add_to_rels(hash_to_hash_rels, document_hash, hash)
            return [hash]

    return []


def process_managed_policy(policy):
    policy_arn = arn.Arn.fromstring(policy['Arn'])
    policy_versions = policy["PolicyVersionList"]
    managed_policy_map[str(policy_arn)] = policy
    policy_version_hashes = process_managed_permission_policy_versions(
        policy_versions)

    for hash in policy_version_hashes:
        add_to_rels(hash_to_arn_rels, hash, str(policy_arn))


def process_inline_policy(inline_policy):
    hash = get_hash(inline_policy)

    if hash in inline_policy_hash_map:
        return hash

    inline_policy['hash'] = hash
    inline_policy_hash_map[hash] = inline_policy

    policy_document_hash = process_permission_document(
        inline_policy['PolicyDocument']
    )

    add_to_rels(hash_to_hash_rels, policy_document_hash, hash)
    return hash


def process_trust_policy(trust_policy: dict):

    trust_policy_hash = get_hash(trust_policy)

    if trust_policy_hash in trust_policy_map:
        print("[*] Trust policy already exists")
        return trust_policy_hash

    trust_policy['hash'] = trust_policy_hash
    trust_policy_map[trust_policy_hash] = trust_policy

    for statement in trust_policy['Statement']:
        statement_hash = process_statement(statement)
        add_to_rels(hash_to_hash_rels, statement_hash, trust_policy_hash)
    return trust_policy_hash


def process_principal_policies(principal: dict):
    principal_arn = principal['Arn']
    for managed_policy in principal['AttachedManagedPolicies']:
        policy_arn = managed_policy['PolicyArn']
        add_to_rels(arn_to_arn_rels, policy_arn, principal_arn)

    process_tags(principal)
    inlines_policy_hashes = []
    key_name = None
    if "RolePolicyList" in principal:
        key_name = "RolePolicyList"
    elif "UserPolicyList" in principal:
        key_name = "UserPolicyList"
    elif "GroupPolicyList" in principal:
        key_name = "GroupPolicyList"

    if key_name:
        for policy in principal[key_name]:
            inline_policy_hash = process_inline_policy(policy)
            inlines_policy_hashes.append(inline_policy_hash)

    for hash in inlines_policy_hashes:
        add_to_rels(hash_to_arn_rels, hash, principal_arn)


def process_user(user):
    user_arn = arn.Arn.fromstring(user['Arn'])
    if str(user_arn) in user_map:
        print("[*] User already processed")
        return

    user_map[str(user_arn)] = user

    process_tags(user)
    process_principal_policies(user)
    for group_name in user["GroupList"]:
        group_arn = get_arn_from_groupname(group_name, user_arn)
        add_to_rels(member_of_rels, user_arn, group_arn)


def process_role(role):
    role_arn = arn.Arn.fromstring(role['Arn'])
    if arn in roles_map:
        print("[*] Role already processed")
        return

    role['rolename'] = role_arn.resource.split("/")[-1]
    roles_map[role_arn] = role

    process_tags(role)
    process_principal_policies(role)
    tp_hash = process_trust_policy(role['AssumeRolePolicyDocument'])
    add_to_rels(hash_to_arn_rels, tp_hash, role_arn)

def get_arn_from_groupname(groupname: str, arn: arn.Arn):
    account_number = arn.account_id
    for group in group_map.values():
        if group['GroupName'] == groupname:
            return group['Arn']

def process_group(group):
    group_arn = arn.Arn.fromstring(group['Arn'])

    if str(group_arn) in group_map:
        print("[*] Group already processed")
        return

    group_map[str(group_arn)] = group

    process_tags(group)
    process_principal_policies(group)


def write_to_csv(filename, items, field_names):
    with open(filename, 'w') as f:
        writer = csv.DictWriter(f, fieldnames=field_names, extrasaction='ignore')
        #writer.writeheader()
        if type(items) == dict:
            items = list(items.values())
        lowercase_items = []
        for item in items:
            lowercase_dict = {}
            for k, v in item.items():
                if type(k) == str:
                    k = k.lower()

                lowercase_dict[k] = v
            lowercase_items.append(lowercase_dict)
        writer.writerows(lowercase_items)


def parse_json(json_text):
    auth_dictionary = json.loads(json_text)
    groups = auth_dictionary["GroupDetailList"]
    users = auth_dictionary["UserDetailList"]
    roles = auth_dictionary["RoleDetailList"]
    policies = auth_dictionary["Policies"]

    account_id = None

    for policy in policies:
        account_id = arn.Arn.fromstring(policy['Arn']).account_id
        process_managed_policy(policy)

    for role in roles:
        process_role(role)

    for group in groups:
        process_group(group)

    for user in users:
        process_user(user)

    


def ingest_csv(session, filename, datatype, fields):
    print(f"[*] Processing csv {filename}")
    query = (
        f'LOAD CSV FROM "file:///{filename}" AS row '
    )

    query += "WITH "

    delimiter = ", "
    field_names = []
    for i in range(len(fields)):
        field_names.append(f"row[{i}] AS {fields[i]}")

    query += delimiter.join(field_names)

    query += (
        f" MERGE (a:{datatype} {{{fields[0]}:{fields[0]}}}) "
        "ON CREATE SET "
    )

    for field in fields:
        query += f"a.{field} = {field}, "
    query += "a.layer = 1 "

    # If the node already exists, delete all properties
    # end reset them
    query += "ON MATCH SET a = {}, "
    for field in fields:
        query += f"a.{field} = {field}, "
    query += "a.layer = 1 "
    try:
        return session.run(query)
    except Exception as e:
        print(e)
        import pdb; pdb.set_trace()


def ingest_resources(session, filename):
    print(f"[*] Processing csv {filename}")
    query = (
        f'LOAD CSV FROM "file:///{filename}" AS row '
        'WITH row[0] AS arn '
        'MERGE (a:UniqueArn {arn:arn}) '
        'ON CREATE SET a.arn = arn, a.layer = 1 '
    )

    session.run(query)



def ingest_relationships(session, filename, source_label, source_field,
                         rel_name, dest_label, dest_field):
    print(f"[*] Processing relationship: {filename}")
    query = (
        f'LOAD CSV FROM "file:///{filename}" AS row '
        'CALL { '
        'WITH row '
        f'MERGE (s:{source_label} {{{source_field}: row[0]}}) '
        f'ON CREATE SET s.inferred = true, s.layer = 1 '
        f'MERGE (d:{dest_label} {{{dest_field}: row[1]}}) '
        f'ON CREATE SET d.inferred = true, d.layer = 1 '
        f'MERGE (s) - [:{rel_name} {{layer: 1}}] -> (d) '
        '} IN TRANSACTIONS'
    )
    session.run(query)


def load_csvs_into_database():
    driver = GraphDatabase.driver("bolt://localhost:7687",
                                  auth=("neo4j", "p@ssw0rd!"))

    with driver.session() as session:
        ingest_csv(session, "actionblobs.csv",
                   "AWSActionBlob:UniqueName",
                   ["name", "regex"])
        ingest_csv(session, "assumerolepolicies.csv", "AWSAssumeRolePolicy:UniqueHash",
                   ["hash", "version", "sid"])
        ingest_csv(session, "conditions.csv", "AWSCondition:UniqueHash",
                   ["hash", "sid"])
        ingest_csv(session, "conditionkeys.csv", "AWSConditionKey:UniqueHash",
                   ["hash", "name"])
        ingest_csv(session, "conditionvalues.csv", "AWSConditionValue:UniqueName",
                   ["name"])
        ingest_csv(session, "groups.csv", "AWSGroup:UniqueArn",
                   ["arn", "path", "name", "groupid",
                    "createdate"])
        ingest_csv(session, "inlinepolicies.csv", "AWSInlinePolicy:UniqueHash",
                   ["hash", "policyname"])

        # The name AWSManagedPolicy is a misnomer because AWS Managed
        # policy implies it is managed by AWS, but a managed policy
        # is one that gets its own ARN. To fit with the naming scheme,
        # we are keeping AWSManagedPolicy
        ingest_csv(session, "managedpolicies.csv", "AWSManagedPolicy:UniqueArn",
                   ["arn", "policyname", "policyid", "path",
                    "defaultversionid", "attachmentcount",
                    "permissionsboundaryusagecount",
                    "isattachable", "createdate", "updatedate"])
        ingest_csv(session, "policydocuments.csv", "AWSPolicyDocument:UniqueHash",
                   ["hash", "version"])
        ingest_csv(session, "policyversions.csv", "AWSPolicyVersion:UniqueHash",
                   ["hash", "versionid", "isdefaultversion"])
        ingest_csv(session, "roles.csv", "AWSRole:UniqueArn",
                   ["arn", "path", "rolename", "roleid", "createdate",
                    "rolelastused"])
        ingest_csv(session, "statements.csv", "AWSStatement:UniqueHash",
                   ["hash", "effect", "sid"])
        ingest_csv(session, "users.csv", "AWSUser:UniqueArn",
                   ["arn", "path", "name", "userid",
                    "createdate"])
        ingest_csv(session, "resourceblobs.csv", "AWSResourceBlob:UniqueName",
                   ['name', 'regex'])
        ingest_csv(session, "tags.csv", "AWSTag:UniqueHash",
                   ['hash', 'key', 'value'])
        ingest_csv(session, "identityproviders.csv", "AWSIdentityProvider:UniqueName",
                   ['name'])
        ingest_csv(session, "principalblobs.csv", "AWSPrincipalBlob:UniqueName", ['name', 'regex'])

        ingest_relationships(session, "hash_to_hash_rels.csv", "UniqueHash",
                             "hash", "AttachedTo", "UniqueHash", "hash")
        ingest_relationships(session, "hash_to_arn_rels.csv", "UniqueHash",
                             "hash", "AttachedTo", "UniqueArn", "arn")
        ingest_relationships(session, "arn_to_arn_rels.csv", "UniqueArn",
                             "arn", "AttachedTo", "UniqueArn", "arn")
        ingest_relationships(session, "member_of_rels.csv", "AWSUser", "arn",
                             "MemberOf", "AWSGroup", "arn")
        ingest_relationships(session, "operator_to_condition_rels.csv",
                             "AWSOperator:UniqueName", "name",
                             "AttachedTo",
                             "AWSCondition:UniqueHash", "hash")
        ingest_relationships(session, "multi_operator_to_condition_rels.csv",
                             "AWSOperator:UniqueName", "name",
                             "AttachedTo",
                             "AWSCondition:UniqueHash", "hash")
        ingest_relationships(session, "statement_to_action_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "Action",
                             "AWSAction:UniqueName", "name")
        ingest_relationships(session, "statement_to_not_action_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "NotAction",
                             "AWSAction:UniqueName", "name")
        ingest_relationships(session, "statement_to_action_blob_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "Action",
                             "AWSActionBlob:UniqueName", "name")
        ingest_relationships(session, "statement_to_not_action_blob_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "NotAction",
                             "AWSActionBlob:UniqueName", "name")
        ingest_relationships(session, "statement_to_resource_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "Resource",
                             "UniqueArn", "arn")
        ingest_relationships(session, "statement_to_not_resource_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "NotResource",
                             "UniqueArn", "arn")
        ingest_relationships(session, "statement_to_resource_blob_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "Resource",
                             "AWSResourceBlob:UniqueName", "name")
        ingest_relationships(session, "statement_to_not_resource_blob_rels.csv",
                             "AWSStatement:UniqueHash", "hash",
                             "NotResource",
                             "AWSResourceBlob:UniqueName", "name")
        
        ingest_relationships(session, "statement_to_principal_arn.csv",
                                "AWSStatement:UniqueHash", "hash",
                                "Principal",
                                "UniqueArn", "arn")
        ingest_relationships(session, "statement_to_principal_uniquename.csv",
                                "AWSStatement:UniqueHash", "hash",
                                "Principal",
                                "UniqueName", "name")
        ingest_relationships(session, "statement_to_principal_blob_rels.csv",
                                "AWSStatement:UniqueHash", "hash",
                                "Principal",
                                "AWSPrincipalBlob:UniqueName", "name")
        
        ingest_relationships(session, "condition_value_to_condition_keys_rels.csv",
                                "AWSConditionValue:UniqueName", "name",
                                "AttachedTo",
                                "AWSConditionKey:UniqueHash", "hash")
                             
        try:
            ingest_resources(session, "arns.csv")
        except neo4j.exceptions.ClientError:
            pass


def rels_to_unique_list(rels):
    return_list = []
    for key, values in rels.items():
        for value in values:
            return_list.append(
                {'source': key,
                 'dest': value})

    return return_list


def write_rels_to_csv(outputdir):
    fields = ['source', 'dest']
    hash_to_hash_filename = os.path.join(outputdir, "hash_to_hash_rels.csv")
    hash_to_arn_filename = os.path.join(outputdir, "hash_to_arn_rels.csv")
    arn_to_arn_rels_filename = os.path.join(outputdir, "arn_to_arn_rels.csv")
    member_of_rels_filename = os.path.join(outputdir, "member_of_rels.csv")

    operator_to_condition_rels_filename = os.path.join(
        outputdir,
        "operator_to_condition_rels.csv")
    multi_operator_to_condition_rels_filename = os.path.join(
        output_dir,
        "multi_operator_to_condition_rels.csv"
    )
    condition_value_to_condition_key_rels_filename = os.path.join(
        outputdir,
        "condition_value_to_condition_keys_rels.csv"
    )
    statement_to_action_rels_filename = os.path.join(
        outputdir,
        "statement_to_action_rels.csv"
    )
    statement_to_not_action_rels_filename = os.path.join(
        outputdir,
        "statement_to_not_action_rels.csv"
    )
    statement_to_resource_rels_filename = os.path.join(
        outputdir,
        "statement_to_resource_rels.csv"
    )
    statement_to_not_resource_rels_filename = os.path.join(
        outputdir,
        "statement_to_not_resource_rels.csv"
    )
    statement_to_resource_blob_rels_filename = os.path.join(
        outputdir,
        "statement_to_resource_blob_rels.csv"
    )
    statement_to_not_resource_blob_rels_filename = os.path.join(
        outputdir,
        "statement_to_not_resource_blob_rels.csv"
    )

    statement_to_action_blob_rels_filename = os.path.join(
        outputdir,
        "statement_to_action_blob_rels.csv"
    )

    statement_to_not_action_blob_rels_filename = os.path.join(
        outputdir,
        "statement_to_not_action_blob_rels.csv"
    )

    condition_key_to_resource_rels_filename = os.path.join(
        outputdir,
        "condition_key_to_resource_rels.csv"
    )

    statement_to_principal_arn_rels_filename = os.path.join(
        outputdir,
        "statement_to_principal_arn.csv"
    )

    statement_to_principal_name_rels_filename = os.path.join(
        outputdir,
        "statement_to_principal_uniquename.csv"
    )

    statement_to_principal_blob_rels_filename = os.path.join(
        outputdir,
        "statement_to_principal_blob_rels.csv"
    )

    condition_value_to_condition_key_rels_filename = os.path.join(
        outputdir,
        "condition_value_to_condition_keys_rels.csv"
    )

    write_to_csv(hash_to_hash_filename,
                 rels_to_unique_list(hash_to_hash_rels), fields)
    write_to_csv(hash_to_arn_filename,
                 rels_to_unique_list(hash_to_arn_rels), fields)
    write_to_csv(arn_to_arn_rels_filename,
                 rels_to_unique_list(arn_to_arn_rels), fields)
    write_to_csv(member_of_rels_filename,
                 rels_to_unique_list(member_of_rels), fields)

    write_to_csv(operator_to_condition_rels_filename,
                 rels_to_unique_list(operator_to_condition_rels),
                 fields)
    write_to_csv(multi_operator_to_condition_rels_filename,
                 rels_to_unique_list(multi_operator_to_condition_rels),
                 fields)
    write_to_csv(statement_to_action_rels_filename,
                 rels_to_unique_list(statement_to_action_rels),
                 fields)
    write_to_csv(statement_to_not_action_rels_filename,
                 rels_to_unique_list(statement_to_not_action_rels),
                 fields)
    write_to_csv(statement_to_resource_rels_filename,
                 rels_to_unique_list(statement_to_resource_rels),
                 fields)
    write_to_csv(statement_to_not_resource_rels_filename,
                 rels_to_unique_list(statement_to_not_resource_rels),
                 fields)
    write_to_csv(statement_to_resource_blob_rels_filename,
                 rels_to_unique_list(statement_to_resource_blob_rels),
                 fields)
    write_to_csv(statement_to_not_resource_blob_rels_filename,
                 rels_to_unique_list(statement_to_not_resource_blob_rels),
                 fields)
    write_to_csv(statement_to_action_blob_rels_filename,
                 rels_to_unique_list(statement_to_action_blob_rels),
                 fields)
    write_to_csv(statement_to_not_action_blob_rels_filename,
                 rels_to_unique_list(statement_to_not_action_blob_rels),
                 fields)
    write_to_csv(condition_key_to_resource_rels_filename,
                 rels_to_unique_list(condition_key_to_resource_rels),
                 fields)
    
    write_to_csv(statement_to_principal_arn_rels_filename,
                    rels_to_unique_list(statement_to_principal_rels),
                    fields)
    write_to_csv(statement_to_principal_name_rels_filename,
                    rels_to_unique_list(statement_to_uniquename_rels),
                    fields)
    
    write_to_csv(statement_to_principal_blob_rels_filename,
                    rels_to_unique_list(statement_to_principal_blob_rels),
                    fields)
    
    write_to_csv(condition_value_to_condition_key_rels_filename,
                    rels_to_unique_list(condition_value_to_key_rels),
                    fields)


def delete_layer_1():
    print("[*] Deleting layer 1")
    driver = GraphDatabase.driver("bolt://localhost:7687",
                                  auth=("bloodhound", "bloodhound"))
    with driver.session() as session:
        for i in [2, 1]:
            session.run(
                f"MATCH () - [r {{layer:{i}}}] - () "
                "DETACH DELETE r"
            )

            session.run(
                f"MATCH (n {{layer: {i}}}) "
                "DELETE n"
            )

            session.run(
                "MATCH (n) - [r] - () WHERE n.layer IS NULL "
                "DETACH DELETE r"
            )

            session.run(
                "MATCH (n) WHERE n.layer IS NULL "
                "DELETE n"
            )


def write_nodes_to_csv(output_dir: str):
    managed_policies_filename = os.path.join(output_dir,
                                             "managedpolicies.csv")

    managed_policies_field_names = ["arn", "policyname", "policyid", "path",
                                    "defaultversionid", "attachmentcount",
                                    "permissionsboundaryusagecount",
                                    "isattachable", "createdate",
                                    "updatedate"]
    write_to_csv(managed_policies_filename, managed_policy_map,
                 managed_policies_field_names)

    policy_version_filename = os.path.join(output_dir, "policyversions.csv")
    policy_version_fields = ['hash', 'versionid', 'isdefaultversion',
                             'createdate']
    write_to_csv(policy_version_filename, policy_version_map,
                 policy_version_fields)

    inline_policy_filename = os.path.join(output_dir, "inlinepolicies.csv")
    inline_poicy_fields = ['hash', 'policyname']
    write_to_csv(inline_policy_filename, inline_policy_hash_map,
                 inline_poicy_fields)

    policy_document_filename = os.path.join(output_dir, "policydocuments.csv")
    policy_document_fields = ['hash', 'version']
    write_to_csv(policy_document_filename, permission_documents_map,
                 policy_document_fields)

    trust_policy_filename = os.path.join(output_dir, "assumerolepolicies.csv")
    trust_policy_fields = ['hash', 'version', 'sid']
    write_to_csv(trust_policy_filename, trust_policy_map, trust_policy_fields)

    statements_filename = os.path.join(output_dir, "statements.csv")
    statement_fields = ['hash', 'effect', 'sid']
    write_to_csv(statements_filename, statements_map, statement_fields)

    conditions_filename = os.path.join(output_dir, "conditions.csv")
    condition_fields = ['hash', 'sid']
    write_to_csv(conditions_filename, condition_map, condition_fields)

    condition_key_filename = os.path.join(output_dir, "conditionkeys.csv")
    condition_key_fields = ['hash', 'name']
    write_to_csv(condition_key_filename, condition_key_map, condition_key_fields)

    condition_value_filename = os.path.join(output_dir, "conditionvalues.csv")
    write_to_csv(condition_value_filename, condition_value_map, ['name'])
    groups_filename = os.path.join(output_dir, "groups.csv")
    group_field_names = ["arn", "path", "groupname", "groupid",
                         "createdate"]
    write_to_csv(groups_filename, group_map, group_field_names)

    roles_filename = os.path.join(output_dir, 'roles.csv')
    role_field_names = ["arn", "path", "rolename", "roleid",
                        "createdate", "rolelastused"]
    write_to_csv(roles_filename, roles_map, role_field_names)

    users_filename = os.path.join(output_dir, "users.csv")
    user_field_names = ["arn", "path", "username", "userid",
                        "createdate"]
    write_to_csv(users_filename, user_map, user_field_names)

    action_blobs_filename = os.path.join(output_dir, "actionblobs.csv")
    write_to_csv(action_blobs_filename, action_blob_map, ['name', "regex"])

    resource_blobs_filename = os.path.join(output_dir, "resourceblobs.csv")
    write_to_csv(resource_blobs_filename, resource_blob_map,
                 ['name', "regex"])

    tags_filename = os.path.join(output_dir, "tags.csv")
    write_to_csv(tags_filename, tag_map, ["hash", "key", "value"])

    idp_filename = os.path.join(output_dir, "identityproviders.csv")
    write_to_csv(idp_filename, identity_provider_map, ["name"])

    principal_blob_filename = os.path.join(output_dir, "principalblobs.csv")
    write_to_csv(principal_blob_filename, principal_blob_map, ["name", "regex"])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-o", "--output-dir",
                        help="The directory to write the output csvs")
    parser.add_argument("-d", "--delete",
                        help="Delete all layer one nodes and relationships",
                        action="store_true")
    parser.add_argument("-i", "--input-dir",
                        help="Input file")
    
    args = parser.parse_args()
    input_dir = args.input_dir
    output_dir = args.output_dir

    if args.delete:
        delete_layer_1()
        sys.exit(0)

    for root, dirs, files in os.walk(input_dir):
        for filename in files:
            if filename.endswith('.json'):
                with open(os.path.join(input_dir, filename), 'r') as f:
                    text = f.read()
                    parse_json(text)
        write_nodes_to_csv(output_dir)
        write_rels_to_csv(output_dir)
        load_csvs_into_database()
