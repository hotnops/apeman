<p align="center">
    <a href="https://join.slack.com/t/bloodhoundhq/shared_invite/zt-1tgq6ojd2-ixpx5nz9Wjtbhc3i8AVAWw">
        <img src="https://img.shields.io/badge/Slack-%23apeman-blueviolet?logo=slack" alt="Slack"/></a>
    <a href="https://twitter.com/hotnops">
        <img src="https://img.shields.io/twitter/url?url=https%3A%2F%2Fx.com%2Fhotnops&style=social"
        alt="@hotnops on Twitter"/></a>
</p>

---

# Project Apeman
![Bigfoot V1](https://github.com/user-attachments/assets/451a052a-97ae-4a95-ab23-f4d3f01ec93f)


# Getting Started
## System Requirements
 - Tested On
    - Windows 11
    - Ubuntu 22
 - 12 GB RAM (This can be reduced in the compose.yaml depending on AWS env size)

## Dependencies

- Docker
- Docker compose
- Python 3
```
sudo apt install python
```
- Python Virtual Environment
```
sudo apt install python3-venv
```

## Running Apeman

### Starting the service

```
git clone git@github.com:hotnops/apeman.git
cd apeman
mkdir import // THIS IS REALLY IMPORTANT
sudo docker compose -f compose.yaml up --build
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

The first time that you start apeman, the AWS nodes and relationships need to be added to the graph. This includes all services, actions, resource types, and condition keys. THIS ONLY NEEDS TO BE RUN ONCE! If AWS updates a service or adds an action, then you will need to re-run this command to honor the new changes. To do this, run the following command:

```
// From apeman/utils
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
```
Optionally, you can obtain a list of all the ARNs in the account. This may help produce more accurate results
```
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

### Reingesting data
If you have updated or new JSON files, you will need to re-ingest all the data and re-analyze. To remove all data, run the following command
```
python -m ingest.ingest -d
```
After this, rerun the ingest and analyze commands:
```
python -m ingest.ingest -i ../path/to/gaad/directory -o ../import
python -m analyze.analyze
```

# Using Apeman

In a browser, navigate to:

```
http://apeman.localhost
```
