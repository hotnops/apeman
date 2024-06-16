#!/usr/bin/env python3
import arn
import re

from neo4j import GraphDatabase


def populate_resource_blob(session):
    print("[*] Expanding resource blobs")
    cypher_query = """
    CALL apoc.periodic.iterate(
        "MATCH (a:AWSResourceBlob) RETURN a",
        "
        MATCH (b:UniqueArn)
        WHERE b.arn =~ a.regex
        MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)
        ",
        {
            batchSize: 10, 
            parallel: true
        }
    ) YIELD batch, failedBatches, errorMessages
    RETURN batch, failedBatches, errorMessages
    """
    result = session.run(cypher_query)
    
    for record in result:
        if record['failedBatches'] > 0:
            print(f"Batch {record['batch']} failed with errors: {record['errorMessages']}")
        else:
            print(f"Batch {record['batch']} processed successfully")


# def populate_not_resources(session):
#     print("[*] Expanding not resource blobs")
#     cypher_query = """
#     CALL apoc.periodic.iterate(
#         "MATCH (s:AWSStatement) - [:NotResource] -> (b:AWSResourceBlob) RETURN s",
#         "
#         MATCH (r:AWSResourceBlob {name: '*'})
#         MERGE (s) - [:Resource {layer: 2}] -> (r)
#         ",
#         {batchSize: 1000, parallel: true}
#     )
#     """
#     session.run(cypher_query)


def populate_action_blob(session):
    print("[*] Expanding action blobs")
    cypher_query = """
    CALL apoc.periodic.iterate(
        "MATCH (a:AWSActionBlob) RETURN a",
        "
        MATCH (b:AWSAction)
        WHERE b.name =~ a.regex
        MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)
        ",
        {batchSize: 10, parallel: true}
    )
    """
    session.run(
        cypher_query
    )

# def populate_not_actions(session):
#     print("[*] Expanding not action blobs")
#     # If the statement uses a notaction, we will simply
#     # point the action blob to a wildcard action. The 
#     # filtering logic will be done post query
#     cypher_query = """
#     CALL apoc.periodic.iterate(
#         "MATCH (s:AWSStatement) - [:NotAction] -> (b:AWSAction|AWSActionBlob) RETURN s",
#         "
#         MATCH (a:AWSAction {name: '*'})
#         MERGE (s) - [:Action {layer: 2}] -> (a)
#         ",
#         {batchSize: 1000, parallel: true}
#     )
#     """
#     session.run(cypher_query)

def convert_variable_arn_to_regex(variable_arn: str) -> str:
    variable_pattern = r'\${[^}]+}'
    variables = re.findall(variable_pattern, variable_arn)
    for variable in variables:
        variable_arn = variable_arn.replace(variable, r'[^:]+')

    return variable_arn



def get_all_arn_nodes(session):
    results = session.run("MATCH (u:UniqueArn) RETURN u")
    return results.values()


def populate_resource_types(session):
    print("[*] Expanding resources types")
    print("[*] Getting all resource types")
    cypher_query = """
    CALL apoc.periodic.iterate(
        "MATCH (a:AWSResourceType) RETURN a",
        "MATCH (b:UniqueArn) WHERE (b.arn =~ a.regex)
         MERGE (b)  - [:TypeOf] -> (a)",
        {batchSize: 10, parallel: true}
        )
    """
    session.run(cypher_query)

def get_allow_assume_role_statements(session):
    print("[*] Getting assume role statements")
    # First, get a pairing of all the allow roles to resources
    print("[*] Creating ALLOW_PERMISSIONS_STS_ASSUMEROLE relationships")
    query = """
        MATCH (prin) <- [:AttachedTo*2..3] - (:AWSPolicyDocument) 
            <- [:AttachedTo] - (s:AWSStatement) 
            - [:Action|ExpandsTo*1..2] -> (a:AWSAction {name: 'sts:assumerole'})
        WHERE (prin:AWSRole OR prin:AWSUser OR prin:AWSGroup)

        WITH prin, s,
            CASE WHEN s.effect = "Allow" THEN s ELSE NULL END as allowStatement,
            CASE WHEN s.effect = "Deny" AND NOT EXISTS ((s) <- [:AttachedTo] - (:AWSCondition)) THEN s ELSE NULL END as denyStatement

        WITH prin,
            COLLECT(allowStatement) as allowStatements,
            COLLECT(denyStatement) as denyStatements

        UNWIND allowStatements as allowSt
        MATCH (allowSt)-[:Resource|ExpandsTo*1..2]->(r:AWSRole)
        WITH DISTINCT r, prin, denyStatements, allowSt
        WHERE NONE(denySt IN denyStatements WHERE (denySt)-[:Resource]->(r) OR (denySt)-[:Resource]->(:AWSResourceBlob)-[:ExpandsTo]->(r))
        MERGE (prin) - [:AllowPermissionSTSAssumeRole {layer: 2, statement:allowSt.hash}] -> (r)
    """
    session.run(query)

    print("[*] Creating ALLOW_TRUST_STS_ASSUMEROLE relationships")
    query = """
        MATCH (prin:AWSRole) <- [:AttachedTo] - (:AWSAssumeRolePolicy) 
            <- [:AttachedTo] - (s:AWSStatement) 
            - [:Action|ExpandsTo*1..2] -> (a:AWSAction {name: 'sts:assumerole'})

        WITH prin, s,
            CASE WHEN s.effect = "Allow" THEN s ELSE NULL END as allowStatement,
            CASE WHEN s.effect = "Deny" AND NOT EXISTS ((s) <- [:AttachedTo] - (:AWSCondition)) THEN s ELSE NULL END as denyStatement

        WITH prin,
            COLLECT(allowStatement) as allowStatements,
            COLLECT(denyStatement) as denyStatements

        UNWIND allowStatements as allowSt

        MATCH (allowSt)-[:Resource|ExpandsTo*1..2]->(r:UniqueArn)
        WITH DISTINCT r, prin, denyStatements, allowSt
        WHERE NONE(denySt IN denyStatements WHERE (denySt)-[:Resource]->(r) OR (denySt)-[:Resource]->(:AWSResourceBlob)-[:ExpandsTo]->(r))
        MERGE (prin) - [:AllowTrustSTSAssumeRole {layer: 2, statement:allowSt.hash}] -> (r)
    """
    session.run(query)
    
    print("[*] Creating CAN_ASSUME relationships")
    query = """
    MATCH (a) - [:AllowPermissionSTSAssumeRole] -> (b) WHERE (b) - [:AllowTrustSTSAssumeRole] -> (a)
    MERGE (a) - [:CanAssume {layer: 2} ] -> (b)
    """
    session.run(query)

    query = """
    MERGE (source) - [:CanAssume {layer: 2} ] -> (dest)
    """
    session.run(query)

    return None


def create_can_assume_edge(statement):

    resources = get_associated_resources(statement)


def rsop_check(roles_to_statements, statements_to_actions):
    rsop_map = {}
    for principal_arn, statement_hashes in roles_to_statements.items():

        prinipal_arn, statement_hash, statement_action = entry
        principal_entry = rsop_map.get(prinipal_arn, {'Allow': [], 'Deny': []})
        principal_entry[statement_action].append(statements_to_actions[statement_hash])
    
    return rsop_map

def query_results_to_dictionary(results):
    ret_dict = {}
    for entry in results:
        hash_list = ret_dict.get(entry[0], [])
        hash_list.append(entry[1])
        ret_dict[entry[0]] = hash_list

    return ret_dict


def get_tier_zero_roles(session):
    # This query gets all principals that have statements that permit
    # Attach and Detach a user role policies to themselves
    query = (
        'MATCH (r:AWSRole) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:attachrolepolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as attachrolestatements, r, statements '
        'MATCH p2=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:detachrolepolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as detachrolestatements, r, attachrolestatements, statements '
        'RETURN DISTINCT r.arn, attachrolestatements, detachrolestatements '
    )

    query_results = session.run(query).values()
    allow_arns = [role_arn[0] for role_arn in query_results]

    query = (
        'MATCH (r:AWSRole) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:DenyAction|ExpandsTo*1..2] -> (a:AWSAction) '
        'WHERE s in statements AND a.name in ["iam:attachrolepolicy", "iam:detachrolepolicy"] '
        'RETURN DISTINCT r.arn, s '
    )

    query_results = session.run(query).values()
    deny_arns = [role_arn[0] for role_arn in query_results]

    t0_roles = [t0_arn for t0_arn in allow_arns if t0_arn not in deny_arns]
    for t0_role in t0_roles:
        query = f'MATCH (a:AWSRole) WHERE a.arn = "{t0_role}" SET a.tier_zero = true'
        query_results = session.run(query)
        print(query_results)
    return t0_roles

    # We now need to check if there is a deny

def get_tier_zero_users(session):
    # This query gets all principals that have statements that permit
    # Attach and Detach a user role policies to themselves
    query = (
        'MATCH (r:AWSUser) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:attachuserpolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as attachstatements, r, statements '
        'MATCH p2=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:detachuserpolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as detachstatements, r, attachstatements, statements '
        'RETURN DISTINCT r.arn, attachstatements, detachstatements '
    )

    query_results = session.run(query).values()
    allow_arns = [user_arn[0] for user_arn in query_results]

    query = (
        'MATCH (r:AWSUser) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:DenyAction|ExpandsTo*1..2] -> (a:AWSAction) '
        'WHERE s in statements AND a.name in ["iam:attachuserpolicy", "iam:detachuserpolicy"] '
        'RETURN DISTINCT r.arn, s '
    )

    query_results = session.run(query).values()
    deny_arns = [user_arn[0] for user_arn in query_results]

    return [t0_arn for t0_arn in allow_arns if t0_arn not in deny_arns]

def get_tier_zero_groups(session):
    # This query gets all principals that have statements that permit
    # Attach and Detach a user role policies to themselves
    query = (
        'MATCH (r:AWSGroup) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:attachgrouppolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as attachstatements, r, statements '
        'MATCH p2=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:Action|ExpandsTo*1..2] -> (:AWSAction {name: "iam:detachgrouppolicy"}) '
        'WHERE s in statements '
        'WITH COLLECT(s) as detachstatements, r, attachstatements, statements '
        'RETURN DISTINCT r.arn, attachstatements, detachstatements '
    )

    query_results = session.run(query).values()
    allow_arns = [group_arn[0] for group_arn in query_results]

    query = (
        'MATCH (r:AWSGroup) <- [:Resource|ExpandsTo*1..2] - (s:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo*1..3] -> (r) '
        'WITH r,COLLECT(s) as statements '
        'MATCH p=(r) <- [:AttachedTo*1..3] - (:AWSPolicyDocument) <- [:AttachedTo] - (s) - [:DenyAction|ExpandsTo*1..2] -> (a:AWSAction) '
        'WHERE s in statements AND a.name in ["iam:attachgrouppolicy", "iam:detachgrouppolicy"] '
        'RETURN DISTINCT r.arn, s '
    )

    query_results = session.run(query).values()
    deny_arns = [user_arn[0] for user_arn in query_results]

    return [t0_arn for t0_arn in allow_arns if t0_arn not in deny_arns]


def get_principals_on_resource(session, resourceArn: str):
    # Get all statements that affect the provided principal
    query = (
        f'MATCH (s:AWSStatement), (u:UniqueArn {{arn: "{resourceArn}"}}) '
        'WHERE (s) - [:Resource] -> (u) OR '
        '(s) - [:Resource] -> (:AWSResourceBlob) - [:ExpandsTo] -> (u) '
        'WITH s,u '
        # Check if the statement actually has an action that can act on the resource
        'MATCH (u) - [:TypeOf] -> (:AWSResourceType) <- [:ActsOn] - (a:AWSAction) <- [:Action|ExpandsTo*1..2] - (s) '
        'WITH COLLECT(s) as statements, a,u '
        'MATCH  (s2:AWSStatement) - [:AttachedTo] -> (:AWSPolicyDocument) - [:AttachedTo|MemberOf*1..4] -> (r) '
        'WHERE s2 IN statements AND (r:AWSRole OR r:AWSUser OR r:AWSGroup) '
        # Typical policy access only affects resources within the same account
        # Empty account ID check is for S3 buckets
        'AND ((r.account_id = u.account_id) OR (u.account_id = "")) '
        'RETURN DISTINCT r.arn,s2.hash,s2.effect,a.name'
    )

    # The returnted statements are any statement that has the target ARN in scope
    # AND has at least one action that affects that resource
    query_result = session.run(query)
    principal_permissions_dict = {}
    for entry in query_result:
        principal_arn, statement_hash, statement_effect, action_name = entry
        rsop = principal_permissions_dict.get(principal_arn, {'Allow': {},
                                                              'Deny': {},
                                                              'ConditionalAllow': {},
                                                              'ConditionalDeny': {}})
        try:
            hash_list = rsop[statement_effect].get(action_name, [])
        except:
            import pdb; pdb.set_trace()
        hash_list.append(statement_hash)
        rsop[statement_effect][action_name] = hash_list
        principal_permissions_dict[principal_arn] = rsop

    return_dict = {}
    for arn, rsop in principal_permissions_dict.items():
        rsop_actions = [x for x in rsop['Allow'].keys() if x not in rsop['Deny'].keys()]
        if rsop_actions:
            action_to_statements = {}
            for rsop_action in rsop_actions:
                action_to_statements[rsop_action] = rsop["Allow"][rsop_action]
            return_dict[arn] = action_to_statements

    return return_dict

def populate_arn_fields(session):
    print("[*] Extrapolating ARN properties")
    # results = session.run("MATCH (u:UniqueArn) RETURN u")
    # for result in results:
    #     try:
    #         node_arn_string = result['u']['arn']
    #         node_arn = arn.Arn.fromstring(node_arn_string)
    #         query = (
    #             f'MATCH (u:UniqueArn {{arn: "{node_arn_string}"}}) '
    #             f'SET u.partition = "{node_arn.partition}" '
    #             f'SET u.service =  "{node_arn.service}" '
    #             f'SET u.region = "{node_arn.region}" '
    #             f'SET u.account_id = "{node_arn.account_id}" '
    #             f'SET u.resource = "{node_arn.resource}" '
    #             f'MERGE (a:AWSAccount {{account_id: "{node_arn.account_id}", layer: 2}}) '
    #             f'MERGE (u) - [:InAccount {{layer: 2}}] -> (a)'
    #         )

    #         session.run(query)
    #     except ValueError:
    #         continue
    cypher_query = """
    CALL apoc.periodic.iterate(
        "MATCH (u:UniqueArn) RETURN u",
        "
        WITH u, apoc.text.regexGroups(u.arn, 'arn:([^:]*):([^:]*):([^:]*):([^:]*):(.+)')[0] AS arn_parts
        WHERE size(arn_parts) = 6
        SET u.partition = arn_parts[1],
            u.service = arn_parts[2],
            u.region = arn_parts[3],
            u.account_id = arn_parts[4],
            u.resource = arn_parts[5]
        RETURN u, arn_parts
        ",
        {batchSize: 1000, parallel: true}
    )
    """
    session.run(cypher_query)


def analyze():
    driver = GraphDatabase.driver("bolt://localhost:7687",
                                  auth=("neo4j", "p@ssw0rd!"))

    with driver.session() as session:
        populate_arn_fields(session)
        populate_resource_types(session)
        populate_action_blob(session)
        #populate_not_actions(session)
        populate_resource_blob(session)
        #populate_not_resources(session)
        #(session)

        # For process each statement and find unconditional
        # sts:assumerole actions and map them to the resources
        # identified in that statement

        #statements = get_allow_assume_role_statements(session)
        #for statement in statements:
        #    create_can_assume_edge(statement)


if __name__ == "__main__":
    analyze()
