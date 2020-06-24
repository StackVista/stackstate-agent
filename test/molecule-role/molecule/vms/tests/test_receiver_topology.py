import json
import os
import re
import util
from molecule.util import safe_load_file
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('receiver_vm')


def _get_instance_config(instance_name):
    instance_config_dict = safe_load_file(os.environ['MOLECULE_INSTANCE_CONFIG'])
    return next(item for item in instance_config_dict if item['instance'] == instance_name)


def _component_data(json_data, type_name, external_id_prefix, command):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
                p["TopologyComponent"]["externalId"].startswith(external_id_prefix):
            component_data = json.loads(p["TopologyComponent"]["data"])
            if command:
                if "args" in component_data["command"]:
                    if component_data["command"]["args"][0] == command:
                        return component_data
            else:
                return component_data
    return None


def _find_component(json_data, type_name, external_id_assert_fn):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
                external_id_assert_fn(p["TopologyComponent"]["externalId"]):
            return p["TopologyComponent"]
    return None


def _relation_data(json_data, type_name, external_id_assert_fn):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyRelation" in p and \
            p["TopologyRelation"]["typeName"] == type_name and \
                external_id_assert_fn(p["TopologyRelation"]["externalId"]):
            return json.loads(p["TopologyRelation"]["data"])
    return None


def _find_process_by_command_args(json_data, type_name, cmd_assert_fn):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
                "data" in p["TopologyComponent"]:
            component_data = json.loads(p["TopologyComponent"]["data"])
            if "args" in component_data["command"] and cmd_assert_fn(' '.join(component_data["command"]["args"])):
                return component_data
    return None


def test_dnat(host, common_vars):
    url = "http://localhost:7070/api/topic/sts_topo_process_agents?offset=0&limit=1000"

    ubuntu_private_ip = _get_instance_config("agent-ubuntu")["private_address"]
    fedora_private_ip = _get_instance_config("agent-fedora")["private_address"]
    dnat_service_port = int(common_vars["dnat_service_port"])
    dnat_server_port = int(common_vars["dnat_server_port"])

    def wait_for_components():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-topo-process-agents-dnat.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        endpoint_match = re.compile("urn:endpoint:/.*:{}".format(ubuntu_private_ip))
        endpoint = _find_component(
            json_data=json_data,
            type_name="endpoint",
            external_id_assert_fn=lambda v: endpoint_match.findall(v))
        assert json.loads(endpoint["data"])["ip"] == ubuntu_private_ip
        endpoint_component_id = endpoint["externalId"]

        proc_to_proc_id_match = re.compile("TCP:/urn:process:/agent-fedora:.*:.*->{}:{}".format(endpoint_component_id, dnat_service_port))
        proc_to_service_id_match = re.compile("TCP:/urn:process:/agent-fedora:.*:.*->urn:process:/agent-ubuntu:.*:.*:{}:{}".format(ubuntu_private_ip, dnat_server_port))
        service_to_proc_id_match = re.compile("TCP:/{}:{}->urn:process:/agent-ubuntu:.*:.*:{}:{}".format(endpoint_component_id, dnat_service_port, ubuntu_private_ip, dnat_server_port))
        assert _relation_data(
            json_data=json_data,
            type_name="directional_connection",
            external_id_assert_fn=lambda v: proc_to_proc_id_match.findall(v))["outgoing"]["ip"] == fedora_private_ip
        assert _relation_data(
            json_data=json_data,
            type_name="directional_connection",
            external_id_assert_fn=lambda v: proc_to_service_id_match.findall(v))["outgoing"]["ip"] == fedora_private_ip
        assert _relation_data(
            json_data=json_data,
            type_name="directional_connection",
            external_id_assert_fn=lambda v: service_to_proc_id_match.findall(v))["incoming"]["ip"] == ubuntu_private_ip

    util.wait_until(wait_for_components, 30, 3)


def test_topology_filtering(host, common_vars):
    url = "http://localhost:7070/api/topic/sts_topo_process_agents?offset=0&limit=2000"

    def wait_for_components():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-topo-process-agents-filtering.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        # assert that we get the stress process and that it contains the top resource tags
        stress_process_match = re.compile("/usr/bin/stress --vm .* --vm-bytes .*")
        stress_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: stress_process_match.findall(v)
        )

        assert stress_process["command"]["exe"] == "/usr/bin/stress"
        assert "usage:top-mem" in stress_process["tags"] or "usage:top-cpu" in stress_process["tags"]

        # assert that we don't get the short-lived python processes
        short_lived_process_match = re.compile("python -c import time; time.sleep(.*);")
        assert _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: short_lived_process_match.findall(v)
        ) is None

    util.wait_until(wait_for_components, 30, 3)
