#!/usr/bin/env python3
from bs4 import BeautifulSoup
import arn

from io import StringIO
import json
import multiprocessing
import neo4j
import os
import pandas as pd
import requests
import shutil
import time

from neo4j import GraphDatabase

NEO4J_BASE = "."
SERVICE_URL = 'https://awspolicygen.s3.amazonaws.com/js/policies.js'
url_base = "https://docs.aws.amazon.com/service-authorization/latest/reference"
services_json_url = f"{url_base}/toc-contents.json"
SERVICE_METADATA_FILE = "service_metadata.json"
MODULE_DIR = os.path.dirname(os.path.abspath(__file__))

def get_service_metadata():
    response = requests.get(SERVICE_URL)
    body_response = str(response.content)
    start_index = body_response.find('{')
    return_dictionary = {}

    metadata = json.loads(body_response[start_index:-1])
    for key, value in metadata['serviceMap'].items():
        key_name = key.replace(" ", "").lower().replace(")", "").replace("(", "")
        return_dictionary[key_name] = value

    return return_dictionary


def get_resource_type(resource_name, resources_data):
    for resource in resources_data:
        if resource_name == resource['Resource types']:
            service_name = resource['ARN'].split(':')[2]
            return f"{service_name}:{resource_name}"


def normalize_dictionary(data, resources_data):
    normalized_dictionary = {}
    for action in data:
        action_name = action['Actions'].replace(" [permission only]", "")
        action_dict = normalized_dictionary.get(action_name, {})
        action_dict['description'] = action['Description']
        if action['Access level']:
            action_dict['access_level'] = action['Access level']
        rts = action_dict.get("resource_types", [])
        rt_name = action.get("Resource types (*required)", None)
        if rt_name:
            rt_name = rt_name.replace("*", "")
            resource_type = get_resource_type(rt_name, resources_data)
            if resource_type:
                rts.append(resource_type)

        condition_keys = action_dict.get("condition_keys", [])
        ck_name = action.get("Condition keys", None)
        if ck_name:
            condition_keys.append(ck_name)

        action_dict['resource_types'] = rts
        normalized_dictionary[action_name] = action_dict

    return normalized_dictionary


def get_table_data(table):
    html_io = StringIO(str(table))
    df = pd.read_html(html_io)[0]
    table_json = df.to_json(orient='records')
    table_data = json.loads(table_json)
    return table_data


def html_tables_to_json(html_content):

    # Parse HTML with BeautifulSoup
    soup = BeautifulSoup(html_content, 'html.parser')
    actions_data = []
    resources_data = []
    conditions_data = []

    # Find the first table in the HTML
    tables = soup.findAll('table')

    for table in tables:
        table_name = table.th.text
        if table_name == "Actions":
            actions_data = get_table_data(table)
        elif table_name == "Resource types":
            resources_data = get_table_data(table)
        elif table_name == "Condition keys":
            conditions_data = get_table_data(table)
        else:
            print(f"[*] Unknown table type: {table_name}")

    return actions_data, resources_data, conditions_data


def extract_hrefs(data):
    href_list = data['contents'][0]['contents'][0]['contents']
    hrefs = list(map(lambda x: x['href'], href_list))
    return hrefs


def get_service_dict(ret_data, link):
    try:
        service_name = link.removesuffix('.html').removeprefix("list_")
        service_name = service_name.replace(" ", "").lower()
        response = requests.get(f"{url_base}/{link}")
        response.raise_for_status()

        html_content = response.text
        actions_json, resources_json, condition_json = html_tables_to_json(
            html_content)
        service_dict = {}
        service_dict['Actions'] = normalize_dictionary(
            actions_json,
            resources_json)
        service_dict['ResourceTypes'] = resources_json
        service_dict['ConditionKeys'] = condition_json
        ret_data[service_name] = service_dict

    except requests.RequestException as e:
        print(f"Failed to fetch content from {link}. Reason: {e}")


def get_all_services_html(data):
    all_hrefs = extract_hrefs(data)
    manager = multiprocessing.Manager()
    ret_data = manager.dict()
    jobs = []
    for link in all_hrefs:
        p = multiprocessing.Process(target=get_service_dict,
                                    args=(ret_data, link))
        jobs.append(p)
        p.start()

    for proc in jobs:
        proc.join()

    return ret_data.copy()


def get_service_prefix(service_name):
    if not os.path.exists(SERVICE_METADATA_FILE):
        print("[*] No existing service metadata file. Building...")
        return_dictionary = get_service_metadata()
        with open(SERVICE_METADATA_FILE, 'w') as f:
            f.write(json.dumps(return_dictionary, sort_keys=True, indent=4))

    with open(SERVICE_METADATA_FILE) as f:
        service_data = json.loads(f.read())
        if service_name not in service_data:
            print(f"[*] Service name: {service_name} not found")
            return None
        return service_data[service_name]["StringPrefix"]


def write_data_to_csv(actions: dict, actions_file, resources_file,
                      condition_keys_file, action_to_resources_types_rels_file):
    actions_file.write("name,access_level\n")
    resources_file.write("name,arn\n")
    action_to_resources_types_rels_file.write("source,dest\n")
    for service_name, definition in actions.items():
        service_prefix = get_service_prefix(service_name)
        lines = []
        for action_name, action_def in definition["Actions"].items():
            full_name = f"{service_prefix}:{action_name}"
            access_level = action_def['access_level']
            lines.append(f"{full_name},{access_level}\n".lower())
            for resource in action_def.get('resource_types', []):
                action_to_resources_types_rels_file.write(
                    f"{full_name},{resource}\n".lower()
                )

        actions_file.writelines(lines)

        lines = []
        resource_types = definition.get("ResourceTypes", [])
        for resource_type in resource_types:
            name = resource_type['Resource types']
            resource_arn = resource_type['ARN']
            rt_service = arn.Arn.fromstring(resource_arn).service
            full_name = f"{rt_service}:{name}"
            lines.append(f"{full_name},{resource_arn}\n".lower())
        resources_file.writelines(lines)

        lines = []
        condition_keys = definition.get("ConditionKeys", [])
        for condition_key in condition_keys:
            name = condition_key['Condition keys']
            lines.append(f"{name}\n".lower())
        condition_keys_file.writelines(lines)


def get_services_json():
    response = requests.get(services_json_url)
    return json.loads(response.text)


def ingest_csv(session, filename, datatype, fields):    
    query = (
        f'LOAD CSV FROM "file:///{filename}" AS row WITH '
    )

    row_names = ", "
    rows = []
    for i in range(len(fields)):
        rows.append(f"row[{i}] as {fields[i]}")

    query += row_names.join(rows)

    query += (
        f" MERGE (a:{datatype} {{{fields[0]}:{fields[0]}}}) "
        "ON CREATE SET "
    )

    for field in fields:
        query += f"a.{field} = {field}, "

    query += "a.layer = 0"
    session.run(query)

def ingest_relationships(session, filename, source_label, source_field,
                         rel_name, dest_label, dest_field):
    print(f"[*] Processing relationship: {filename}")
    query = (
        f'LOAD CSV FROM "file:///{filename}" AS row '
        'CALL { '
        'WITH row '
        f'MERGE (s:{source_label} {{{source_field}: row[0]}}) '
        f'ON CREATE SET s.inferred = true '
        f'MERGE (d:{dest_label} {{{dest_field}: row[1]}}) '
        f'ON CREATE SET d.inferred = true '
        f'MERGE (s) - [:{rel_name} {{layer: 0}}] -> (d) '
        '} IN TRANSACTIONS'
    )
    session.run(query)


def load_csvs_into_database(driver):


    with driver.session() as session:
        ingest_csv(session, "awsmultivaluedprefix.csv",
                   "AWSMultivalueOperator:UniqueName",
                   ["name"])
        ingest_csv(session, "awsoperators.csv", "AWSOperator:UniqueName",
                   ["name"])
        ingest_csv(session, "awsresourcetypes.csv", "AWSResourceType:UniqueName",
                   ["name", "arn"])
        ingest_csv(session, "awsglobalconditionkeys.csv", "AWSConditionKey:UniqueName",
                   ["name"])
        ingest_csv(session, "awsconditionkeys.csv", "AWSConditionKey:UniqueName",
                   ["name"])
        ingest_csv(session, "awsactions.csv", "AWSAction:UniqueName",
                   ["name", "access_level"])
        ingest_relationships(session, "actions_to_resourcetypes_rels.csv", "AWSAction:UniqueName",
                             "name", "ActsOn", "AWSResourceType:UniqueName", "name")

def create_constraint(session, constraint_name, label, property):    
    query = (
        f"CREATE CONSTRAINT {constraint_name} IF NOT EXISTS "
        f"FOR (n:{label}) REQUIRE n.{property} IS UNIQUE")

    session.run(query)

def create_relationship_constraint(session, constraint_name, rel_name,
                                   unique_property_name):
    query = (
        f"CREATE CONSTRAINT {constraint_name} IF NOT EXISTS "
        f"FOR () - [r:{rel_name}] -() REQUIRE (r.{unique_property_name}) IS UNIQUE"
    )
    session.run(query)


def create_constraints(driver):
    with driver.session() as session:
        create_constraint(session, 'awsactionconstraint',
                          'AWSAction', 'name')
        create_constraint(session, 'actionblobconstraint',
                          'AWSActionBlob', 'name')
        create_constraint(session, 'assumerolepolicyconstraint',
                          'AWSAssumeRolePolicy', 'hash')
        create_constraint(session, "conditionconstraint",
                          "AWSCondition", "hash")
        create_constraint(session, "conditionvalueconstraint",
                          "AWSConditionValue", "name")
        create_constraint(session, "groupconstraint",
                          "AWSGroup", "arn")
        create_constraint(session, "inlinepolicyconstraint",
                          "AWSInlinePolicy", "hash")
        create_constraint(session, "managedpolicyconstraint",
                          "AWSManagedPolicy", "arn")
        create_constraint(session, "policydocumentconstraint",
                          "AWSPolicyDocument", "hash")
        create_constraint(session, "policyversionconstraint",
                          "AWSPolicyVersion", "hash")
        create_constraint(session, "roleconstraint",
                          "AWSRole", "arn")
        create_constraint(session, "statementconstraint",
                          "AWSStatement", "hash")
        create_constraint(session, "userconstraint",
                          "AWSUser", "arn")
        create_constraint(session, "resourceblobconstraint",
                          'AWSResourceBlob', 'name')
        create_constraint(session, "tagconstraint",
                          'AWSTag', 'hash')

def create_indices(driver):
    with driver.session() as session:
        query = (
            'CREATE TEXT INDEX uniquehash IF NOT EXISTS '
            'FOR (n:UniqueHash) ON (n.hash)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX uniquearn IF NOT EXISTS '
            'FOR (n:UniqueArn) ON (n.arn)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX uniquename IF NOT EXISTS '
            'FOR (n:UniqueName) ON (n.name)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX statementeffect IF NOT EXISTS '
            'FOR (s:AWSStatement) ON (s.effect)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX actionname IF NOT EXISTS '
            'FOR (s:AWSAction) ON (s.name)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX actionblobname IF NOT EXISTS '
            'FOR (s:AWSActionBlob) ON (s.name)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX roleblobname IF NOT EXISTS '
            'FOR (s:AWSResourceBlob) ON (s.name)'
        )
        session.run(query)
        query = (
            'CREATE TEXT INDEX rolebname IF NOT EXISTS '
            'FOR (s:AWSResource) ON (s.name)'
        )
        session.run(query)


def aws_initialize(output_dir, schema_path=None):

    if not schema_path:
        services_json = get_services_json()
        aws_schema = get_all_services_html(services_json)
        print("[*] Writing new schema to awschema.json")
        with open('awsschema.json', 'w') as f:
            f.write(json.dumps(aws_schema, indent=4))
    else:
        try:
            with open('awsschema.json', 'r') as f:
                print("[*] Loading existing schema")
                aws_schema = json.loads(f.read())
        except FileNotFoundError:
            print("[!] Schema file not found. Rerun without input param to generate new schema file")

    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    actions_filename = os.path.join(output_dir, "awsactions.csv")
    resources_filename = os.path.join(output_dir, "awsresourcetypes.csv")
    ck_filename = os.path.join(output_dir, "awsconditionkeys.csv")
    rt_rels_filename = os.path.join(output_dir, "actions_to_resourcetypes_rels.csv")

    with open(
            actions_filename, 'w') as actions_file, open(
            resources_filename, 'w') as resources_file, open(
            ck_filename, 'w') as ck_file, open(
                rt_rels_filename, 'w') as action_to_rt_rels_file:
        write_data_to_csv(aws_schema, actions_file, resources_file,
                          ck_file, action_to_rt_rels_file)

    # These files are manually maintained
    shutil.copy(os.path.join(MODULE_DIR, "awsoperators.csv"),
                os.path.join(output_dir, "awsoperators.csv"))
    shutil.copy(os.path.join(MODULE_DIR, "awsglobalconditionkeys.csv"),
                os.path.join(output_dir, "awsglobalconditionkeys.csv"))
    shutil.copy(os.path.join(MODULE_DIR, "awsmultivaluedprefix.csv"),
                os.path.join(output_dir, "awsmultivaluedprefix.csv"))

    while (True):
        try:
            driver = GraphDatabase.driver("bolt://localhost:7687", 
                                          auth=None)
            load_csvs_into_database(driver)
            create_constraints(driver)
            create_indices(driver)
            break
        except neo4j.exceptions.ServiceUnavailable:
            print("[*] Failed to connect retrying in 5 seconds")
            time.sleep(5)


if __name__ == "__main__":
    import argparse
    argParser = argparse.ArgumentParser()
    argParser.add_argument("-o", "--output-directory",
                           help="The directory to output csv files", required=True)
    argParser.add_argument("-i", "--input-schema", help="The AWS input schema", 
                           default=None)
    args = argParser.parse_args()
    aws_initialize(args.output_directory)
    print("[*] Finished")
