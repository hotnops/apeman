#!/usr/bin/env python3
import arn
import re
import requests

from neo4j import GraphDatabase

def populate_principal_blob(session):  
    print("[*] Expanding principal blobs")
    cypher_query = """
    CALL apoc.periodic.iterate(
        "MATCH (a:AWSPrincipalBlob) RETURN a",
        "
        MATCH (b:AWSUser|AWSRole|AWSGroup|AWSIdentityProvider|AWSService)
        WHERE b.arn =~ a.regex OR b.name =~ a.regex
        MERGE (a) - [:ExpandsTo {layer: 2}] -> (b)
        ",
        {batchSize: 10, parallel: false}
    )
    """
    session.run(cypher_query)

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
        {batchSize: 10, parallel: false}
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
         MERGE (b)  - [:TypeOf {layer: 2}] -> (a)",
        {batchSize: 10, parallel: true}
        )
    """
    session.run(cypher_query)


def populate_arn_fields(session):
    print("[*] Extrapolating ARN properties")
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

def analyze_assume_roles():
    print("[*] Analyzing assume roles")
    resp = requests.get("http://apeman-backend.localhost/analyze/assumeroles")
    if resp.status_code == 200:
        print("[*] Assume role analysis complete")
    else:
        print("[!] Could not analyze assume roles")
        import pdb; pdb.set_trace()

def analyze():
    driver = GraphDatabase.driver("bolt://localhost:7687",
                                  auth=("neo4j", "p@ssw0rd!"))

    with driver.session() as session:
        populate_arn_fields(session)
        populate_resource_types(session)
        populate_action_blob(session)
        #populate_not_actions(session)
        populate_resource_blob(session)
        populate_principal_blob(session)
        #populate_not_resources(session)

        # This is the first migration into server
        # based analysis. All others will move over
        # to API calls
        analyze_assume_roles()


if __name__ == "__main__":
    analyze()
