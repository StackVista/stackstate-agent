import util
import integration_sample
from conftest import STS_CONTEXT_FILE

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def kubernetes_event_data(event, json_data):
    for message in json_data["messages"]:
        p = message["message"]
        if "GenericEvent" in p:
            _data = p["GenericEvent"]
            if _data == dict(_data, **event):
                return _data
    return None


# def test_agent_integration_sample_metrics(cliv1, hostname):
#     # Suspect where formerly sts_multi_metrics topic was merely empty, now it's not even there at all anymore.
#     # expected = {'system.cpu.usage', 'location.availability', '2xx.responses', '5xx.responses', 'check_runs'}
#     expected = None
#     util.assert_metrics(cliv1, hostname, expected)


def test_agent_integration_sample_topology(cliv1):
    expected_components = integration_sample.get_agent_integration_sample_expected_topology()
    util.assert_topology(cliv1, "sts_topo_agent_integrations", expected_components)


def test_agent_integration_sample_topology_events(cliv1):
    expected_topology_events = [
        {
            "assertion": "find the URL timeout topology event",
            "event": {
               "category": "my_category",
               "name": "URL timeout",
               "tags": [],
               "data": "{\"another_thing\":1,\"big_black_hole\":\"here\",\"test\":{\"1\":\"test\"}}",
               "source_identifier": "source_identifier_value",
               "source": "source_value",
               "element_identifiers": [
                   "urn:host:/123"
               ],
               "source_links": [
                   {
                       "url": "http://localhost",
                       "name": "my_event_external_link"
                   }
               ],
               "type": "HTTP_TIMEOUT",
               "description": "Http request to http://localhost timed out after 5.0 seconds."
            }
        }
    ]
    util.assert_topology_events(cliv1, "agent-integration-sample", "sts_topology_events", expected_topology_events)
