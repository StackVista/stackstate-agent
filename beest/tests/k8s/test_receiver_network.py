import json
import re
import util
from conftest import STS_CONTEXT_FILE

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def _find_component(json_data, type_name, external_id_assert_fn):
    messages = json_data["messages"]
    messages.reverse()
    for message in messages:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
                external_id_assert_fn(p["TopologyComponent"]["externalId"]):
            return p["TopologyComponent"]
    return None


def _relation_data(json_data, type_name, external_id_assert_fn):
    messages = json_data["messages"]
    messages.reverse()
    for message in messages:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyRelation" in p and \
            p["TopologyRelation"]["typeName"] == type_name and \
                external_id_assert_fn(p["TopologyRelation"]["externalId"]):
            return json.loads(p["TopologyRelation"]["data"])
    return None


def _find_process_by_command_args(json_data, type_name, cmd_assert_fn):
    messages = json_data["messages"]
    messages.reverse()
    for message in messages:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
                "data" in p["TopologyComponent"]:
            component_data = json.loads(p["TopologyComponent"]["data"])
            if "args" in component_data["command"] and cmd_assert_fn(' '.join(component_data["command"]["args"])):
                return component_data
    return None


limit = 3000


def test_dnat(host, ansible_var, cliv1):
    server_port = int(ansible_var("dnat_server_port"))
    service_port = int(ansible_var("dnat_service_port"))
    global limit
    limit = 3000

    def wait_for_components():
        global limit
        json_data = cliv1.topic_api("sts_topo_process_agents", limit=limit, config_location=STS_CONTEXT_FILE)
        message_count = len(json_data["messages"])
        if message_count >= limit:
            limit += 500

        server_process_match = re.compile("ncat -vv --broker --listen -p {}".format(server_port))
        server_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: server_process_match.findall(v)
        )
        assert server_process is not None
        server_process_create_time = server_process["createTime"]
        server_process_pid = server_process["pid"]
        server_host = server_process["host"]

        request_process_match = re.compile("nc -vv pod-service {}".format(service_port))
        request_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: request_process_match.findall(v)
        )
        assert request_process is not None
        request_process_create_time = request_process["createTime"]
        request_process_pid = request_process["pid"]
        request_host = request_process["host"]

        request_process_to_server_relation_match = re.compile(
            "TCP:/urn:process:/{}:{}:{}->urn:process:/{}:{}:{}:{}"
            .format(request_host, request_process_pid, request_process_create_time,
                    server_host, server_process_pid, server_process_create_time,
                    server_port)
        )

        assert _relation_data(
            json_data=json_data,
            type_name="directional_connection",
            external_id_assert_fn=lambda v: request_process_to_server_relation_match.findall(v)
        ) is not None

    util.wait_until(wait_for_components, 120, 3)


def test_pod_container_to_container(ansible_var, cliv1):
    server_port = int(ansible_var("container_to_container_server_port"))
    global limit
    limit = 3000

    def wait_for_components():
        global limit

        json_data = cliv1.topic_api("sts_topo_process_agents", limit=limit, config_location=STS_CONTEXT_FILE)
        message_count = len(json_data["messages"])
        if message_count >= limit:
            limit += 500

        server_process_match = re.compile("nc -l -p {}".format(server_port))
        server_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: server_process_match.findall(v)
        )
        assert server_process is not None
        server_process_create_time = server_process["createTime"]
        server_process_pid = server_process["pid"]
        server_host = server_process["host"]

        request_process_match = re.compile("nc localhost {}".format(server_port))
        request_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: request_process_match.findall(v)
        )
        assert request_process is not None
        request_process_create_time = request_process["createTime"]
        request_process_pid = request_process["pid"]
        request_host = request_process["host"]

        request_process_to_server_relation_match = "TCP:/urn:process:/{}:{}:{}->urn:process:/{}:{}:{}:{}".format(
            request_host, request_process_pid, request_process_create_time,
            server_host, server_process_pid, server_process_create_time,
            server_port
        )

        assert _relation_data(
                json_data=json_data,
                type_name="directional_connection",
                external_id_assert_fn=lambda v: re.compile(request_process_to_server_relation_match).findall(v)
            ) is not None

    util.wait_until(wait_for_components, 420, 3)


def test_headless_pod_to_pod(ansible_var, cliv1):
    # Server and service port are equal
    server_port = int(ansible_var("headless_service_port"))
    global limit
    limit = 3000

    def wait_for_components():
        global limit
        json_data = cliv1.topic_api("sts_topo_process_agents", limit=limit, config_location=STS_CONTEXT_FILE)
        message_count = len(json_data["messages"])
        if message_count >= limit:
            limit += 500

        server_process_match = re.compile("ncat -vv --broker --listen -p {}".format(server_port))
        server_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: server_process_match.findall(v)
        )
        assert server_process is not None
        server_process_create_time = server_process["createTime"]
        server_process_pid = server_process["pid"]
        server_host = server_process["host"]

        request_process_match = re.compile("nc -vv headless-service {}".format(server_port))
        request_process = _find_process_by_command_args(
            json_data=json_data,
            type_name="process",
            cmd_assert_fn=lambda v: request_process_match.findall(v)
        )
        assert request_process is not None
        request_process_create_time = request_process["createTime"]
        request_process_pid = request_process["pid"]
        request_host = request_process["host"]

        request_process_to_server_relation_match = re.compile(
            "TCP:/urn:process:/{}:{}:{}->urn:process:/{}:{}:{}:{}"
            .format(request_host, request_process_pid, request_process_create_time,
                    server_host, server_process_pid, server_process_create_time,
                    server_port)
        )

        assert _relation_data(
                json_data=json_data,
                type_name="directional_connection",
                external_id_assert_fn=lambda v: request_process_to_server_relation_match.findall(v)
            ) is not None

    util.wait_until(wait_for_components, 120, 3)
