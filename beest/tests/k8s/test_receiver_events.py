import json
import util

testinfra_hosts = ["local"]


def test_generic_events(cliv1):
    def wait_for_events():
        json_data = cliv1.topic_api("sts_generic_events")
        with open("./topic-sts-generic-events.json", 'w') as f:
            json.dump(json_data, f, indent=4)

    util.wait_until(wait_for_events, 60, 3)


def test_topology_events(cliv1):
    def wait_for_topology_events():
        json_data = cliv1.topic_api("sts_topology_events")
        with open("./topic-sts-topology-events.json", 'w') as f:
            json.dump(json_data, f, indent=4)

    util.wait_until(wait_for_topology_events, 60, 3)
