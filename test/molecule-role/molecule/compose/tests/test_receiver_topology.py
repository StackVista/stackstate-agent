import json
import os
import util
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('trace-java-demo')


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


def _relation_data(json_data, type_name, external_id_assert_fn):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyRelation" in p and \
            p["TopologyRelation"]["typeName"] == type_name and \
                external_id_assert_fn(p["TopologyRelation"]["externalId"]):
            return json.loads(p["TopologyRelation"]["data"])
    return None


def test_receiver_ok(host):
    def assert_healthy():
        c = "curl -s -o /dev/null -w \"%{http_code}\" http://localhost:7077/health"
        assert host.check_output(c) == "200"

    util.wait_until(assert_healthy, 30, 3)


def test_agent_ok(host):
    def assert_healthy():
        c = "docker inspect ubuntu_stackstate-agent_1 |  jq -r '.[0].State.Health.Status'"
        assert host.check_output(c) == "healthy"

    util.wait_until(assert_healthy, 30, 3)


def test_java_traces(host):
    def assert_ok():
        c = "curl -H Host:stackstate-books-app -s -o /dev/null -w \"%{http_code}\" http://localhost/stackstate-books-app/listbooks"
        assert host.check_output(c) == "200"

    util.wait_until(assert_ok, 120, 10)

    def assert_components():
        topo_url = "http://localhost:7070/api/topic/sts_topo_process_agents?limit=5000"
        data = host.check_output("curl \"%s\"" % topo_url)
        json_data = json.loads(data)
        with open("./topic-topo-process-agents-traces.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        assert _component_data(json_data, "service", "urn:service:/traefik:stackstate-authors-app", None)["name"] == "traefik:stackstate-authors-app"
        assert _component_data(json_data, "service", "urn:service:/traefik:stackstate-books-app", None)["name"] == "traefik:stackstate-books-app"

        # TODO
        # traefik service
        # postgres db service
        # books app service instance + processes (due to scale)
        # authors app service instance + processes (due to scale)

        # books app service -> instances
        # authors app service -> instances
        # books app service -> traefik
        # authors app service -> traefik
        # traefik -> books app service
        # traefik -> authors app service
        # books app service -> authors app service
        # app service -> postgres service
        # app service instances -> postgres service

    util.wait_until(assert_components, 30, 3)
