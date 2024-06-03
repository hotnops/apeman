# Getting Started

## Dependencies

- Docker
- Docker compose

## Running Apeman

### Starting the service

```
git clone git@github.com:hotnops/apeman.git
cd apeman
sudo docker compose -f compose.yaml up
```

### Creating a venv

A python virtual environment is recommended to ingest the data

```
cd utils
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Initializing the AWS schema

The first time that you start apeman, the AWS nodes and relationships need to be added to the graph. This includes all services, actions, resource types, and condition keys. To do this, run the following command:

```
cd utils
python -m init.aws_initialize -o ../import
```

If initialization is successful, you will get an output like this

```
[*] Writing new schema to awschema.json
[*] Service name: awsre_postprivate not found
[*] Service name: amazonsimpleemailservice-mailmanager not found
[*] Service name: awsusersubscriptions not found
[*] Processing relationship: actions_to_resourcetypes_rels.csv
[*] Finished
```

### Obtaining data

Data needs to be ingested and analyzed before Apeman can present useful information. First, create a new directory
called "gaad"

```
// cd back to the root apeman dir
cd ../
mkdir gaad
touch gaad/arns.json
```

Next, for every account you want analyzed, perform the following action

```
aws iam get-account-authorization-details > gaad/<account_number>.json
aws resource-explorer-2 search --query-string "*" | jq -r '.Resources[] | [.Arn] | @csv' >> import/arns.csv
```

### Ingest the data

Now all the data collected gets ingested into the graph database

```
python -m ingest.ingest -i ../gaad -o ../import
```

If ingest is successful, you will get an output like this:

```
...
[*] Processing relationship: member_of_rels.csv
[*] Processing relationship: condition_key_to_condition_rels.csv
[*] Processing relationship: condition_value_to_condition_rels.csv
[*] Processing relationship: operator_to_condition_rels.csv
[*] Processing relationship: multi_operator_to_condition_rels.csv
[*] Processing relationship: allow_action_to_statement_rels.csv
[*] Processing relationship: deny_action_to_statement_rels.csv
[*] Processing relationship: statement_to_allow_action_blob_rels.csv
[*] Processing relationship: statement_to_deny_action_blob_rels.csv
[*] Processing relationship: statement_to_resource_rels.csv
[*] Processing relationship: statement_to_resource_blob_rels.csv
[*] Processing csv arns.csv
```

### Analyze the data

Lastly, the data needs to be analyzed

```
python -m analyze.analyze
```

# Using Apeman

In a browser, navigate to:

```
http://apeman.localhost
```