#!/usr/bin/env python3

import json
import requests

SERVICE_URL = 'https://awspolicygen.s3.amazonaws.com/js/policies.js'


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


if __name__ == "__main__":
    return_dictionary = get_service_metadata()
    with open('service_metadata.json', 'w') as f:
        f.write(json.dumps(return_dictionary, sort_keys=True, indent=4))
