import json
import subprocess
from box import Box


# Topology script documentation: https://docs.stackstate.com/develop/reference/scripting/script-apis/topology
# sts script run --json --file query.sts | less


def query_components(query):
    script = f"""
    Topology.query('{query}')
    .components()
    """
    return run_script(script)["result"]["value"]

def query_relations(query):
    script = f"""
    Topology.query('{query}')
    .relations()
    """
    return run_script(script)["result"]["value"]


def query_full_components(query):
    script = f"""
    Topology.query('{query}')
    .fullComponents()
    """
    return run_script(script)["result"]["value"]


def run_script(script):
    stdout = subprocess.run(["sts", "script", "run", "--json", "--script", script], capture_output=True).stdout
    return json.loads(stdout)


def assert_topology(expected):
    # print(expected)
    actual_components = query_components(expected['query'])
    # print(json.dumps(actual_components, indent=4))
    expected_components = expected['topology']['components']
    assert len(actual_components) == len(expected_components), f"number of components retrieved ({len(actual_components)}) is not the same as expected ({len(expected_components)})"
    assert_components(expected_components, actual_components)
    print("All good!")



def assert_components(expected, actual):
    merged_components = {}
    for ekey, ecomponent in expected.items():
        match = False
        for acomponent in actual:
            if "name" in ecomponent:
                if ecomponent["name"] != acomponent["name"]:
                    continue
            if "tags" in ecomponent:
                if not assert_tags(ecomponent["tags"], acomponent["tags"]):
                    continue
            # TODO check identifiers
            # TODO check tags
            match = True

        if not match:
            print(f"Component not found: {json.dumps(ecomponent, indent=4)}")
            assert False, f"component {ecomponent.name} not found"

    return merged_components


def assert_tags(expected, actual):
    for etag in expected:
        match = False
        for atag in actual:
            if etag == atag:
                match = True
                break
        if not match:
            return False
    return True


# query = 'label in ("aws-stepfunctions-statemachine", "aws-stepfunctions-state")'
# print(json.dumps(query_components(query)[0], indent=4))
# print(json.dumps(client.query_by_label()[0], indent=4))

# compa = get_component_by_name(A)
# compb = get_component_by_labels(la, lb)
# assert relation(compa, compb)


components = Box({
    "work_on_case": {
        "name": "Work on Case",
        "identifiers_regex": [r"arn:aws:states:.*:statemachine:sts_xray_test_call_center:state/work on case"],
        "tags": ["aws-stepfunctions-state"],
        "data": {},
    },
    "escalate_case": {
        "name": "Escalate Case",
    },
    "sts_xray_test_call_center": {
        "name": "sts_xray_test_call_center",
    },
    "Is Case Resolved": {
        "name": "Is Case Resolved",
    },
    "Open Case": {
        "name": "Open Case",
    },
    "Fail": {
        "name": "Fail",
    },
    "Assign Case": {
        "name": "Assign Case",
    },
    "Close Case": {
        "name": "Close Case",
    }
})
'''
components = Box({
    "work_on_case": {
        "name": "Work on Case",
        "externalId": "",
        "tags": [],
        "data": {},
    },
    "compb": {
        "name": "",
        "type": "",
        "externalId": "",
        "labels": [],
        "data": {},
    }
})
'''
relations = [
    {
        "source": components.work_on_case,
        "target": components.escalate_case,
        "type": "runs on",
        "data": {}
    }
]

expected = {
    "query": 'label in ("aws-stepfunctions-statemachine", "aws-stepfunctions-state")',
    "topology": {
        "components": components,
        "relations": relations
    }
}

# print(json.dumps(query_relations('label in ("aws-stepfunctions-statemachine", "aws-stepfunctions-state")'), indent=4))
assert_topology(expected)


# if _data == dict(_data, **event):
