import util
import json
from util import assert_metrics_check_instance, match_partial_event
from ststest import TopicTopologyMatcher

testinfra_hosts = ["local"]
test_component = "agent_integration_sample_v2_transactional"
check_type = "agent-v2-integration-transactional"
check_url = "sample"
check_identifier = f"{check_type}_{check_url}"


def test_agent_sample_integration_generic_events(cliv1):

    def wait_for_events():
        json_data = cliv1.topic_api("sts_generic_events")
        with open(f"./topic-{test_component}-sts-generic-events.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        service_event = {
            "message": "agent_integration_sample_v2_transactional check was processed successfully",
            "name": "service-check.service-check",
            "title": "agent_integration_sample_v2_transactional",
            "eventType": "service-check",
            "tags": {
                "integration-type": "agent-v2-integration-transactional",
                "integration-url": "sample",
                "source_type_name": "service-check",
                "status": "OK"
            },
        }
        assert match_partial_event(service_event, json_data), f"no matches found for event: {service_event}"

        http_event = {
            "name": "HTTP_TIMEOUT",
            "title": f"URL timeout - {check_identifier}",
            "eventType": "HTTP_TIMEOUT",
            "tags": {
                "source_type_name": "HTTP_TIMEOUT"
            },
            "message": "Http request to http://localhost timed out after 5.0 seconds."
        }
        assert match_partial_event(http_event, json_data), f"no matches found for event: {service_event}"

    util.wait_until(wait_for_events, 60, 5)


def test_agent_integration_sample_metrics(host, cliv1):
    expected = {'system.cpu.usage', 'location.availability', '2xx.responses', '5xx.responses', 'check_runs',
                f'{check_identifier}_persistent_key', f'{check_identifier}_stateful',
                f'{check_identifier}_transactional'}
    json_data = cliv1.topic_api("sts_multi_metrics")

    with open(f"./topic-{test_component}-sts-metrics.json", 'w') as f:
        json.dump(json_data, f, indent=4)

    assert_metrics_check_instance(host, json_data, expected, check_identifier)


def test_agent_integration_sample_topology_events(host, cliv1):

    def wait_for_topology_events():
        json_data = cliv1.topic_api("sts_topology_events")
        with open(f"./topic-{test_component}-sts-topology-events.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        def _topology_event_data(event):
            for message in json_data["messages"]:
                p = message["message"]
                if "TopologyEvent" in p:
                    _data = p["TopologyEvent"]
                    if _data == dict(_data, **event):
                        return _data
            return None

        assert _topology_event_data(
            {
                "category": "my_category",
                "name": f"URL timeout - {check_identifier}",
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
        ) is not None

    util.wait_until(wait_for_topology_events, 60, 3)


def test_agent_integration_sample_health_synchronization(host, cliv1):

    def wait_for_health_messages():
        json_data = cliv1.topic_api("sts_intake_health", 100)
        with open(f"./topic-{test_component}-sts-health-messages.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        def _health_contains_payload(event):
            for message in json_data["messages"]:
                p = message["message"]
                if "IntakeHealthMessage" in p:
                    _data = p["IntakeHealthMessage"]["payload"]
                    if _data == dict(_data, **event):
                        return _data
            return None

        assert _health_contains_payload({
            "IntakeHealthMainStreamStart": {
                "repeatIntervalMs": 10000
            }
        }
        ) is not None
        assert _health_contains_payload({
            "IntakeHealthMainStreamStop": {}
        }
        ) is not None

        data = {
            'checkStateId': 'id',
            'health': 'CRITICAL',
            'message': 'msg',
            'name': 'name',
            'topologyElementIdentifier': check_identifier
        }
        assert _health_contains_payload(
            {
                "IntakeHealthCheckStates": {
                    "consistencyModel": "REPEAT_SNAPSHOTS",
                    "intakeCheckStates": [
                        {"data": json.dumps(data, separators=(',', ':'))}
                    ]
                }
            }
        ) is not None

    util.wait_until(wait_for_health_messages, 60, 3)


def test_agent_integration_sample_topology_topic_api(host, agent_hostname, cliv1):

    agent_integration_sample_topology = TopicTopologyMatcher()\
        .component("this-host-assertion", name=r"this-host", domain=r"Webshop")\
        .component("some-application-assertion", name=r"some-application")\
        .delete(r"urn:example:/host:host_for_deletion")

    def assert_topology():
        topology_result = cliv1.topology_topic(topic=f"sts_topo_{check_type}_{check_url}", limit=20)

        match_result = agent_integration_sample_topology.find(topology_result)
        match_result.assert_exact_match()

    util.wait_until(assert_topology, 60, 3)


def test_agent_integration_transactional_stateful_increment(host, agent_hostname, cliv1):
    json_data = cliv1.topic_api("sts_multi_metrics")

    with open(f"./topic-{test_component}-sts-metrics-transactional-stateful.json", 'w') as f:
        json.dump(json_data, f, indent=4)

    def wait_for_metrics():
        persistent_metric = f'{check_identifier}_persistent_key'
        persistent_metrics_values = []
        stateful_metric = f'{check_identifier}_stateful'
        stateful_metric_values = []
        transactional_metric = f'{check_identifier}_transactional'
        transactional_metric_values = []

        def get_keys(m_host):
            host_metrics = sorted(list(
                {'timestamp': message["message"]["MultiMetric"]["timestamp"],
                 'metrics': message["message"]["MultiMetric"]["values"]}
                for message in json_data["messages"]
                if message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                message["message"]["MultiMetric"]["host"] == m_host and
                check_identifier in message["message"]["MultiMetric"]["labels"]
            ), key=lambda m: m['timestamp'], reverse=True)

            for converted_metric in host_metrics:
                for metric_key, metric_value in converted_metric["metrics"].items():
                    if metric_key == persistent_metric:
                        persistent_metrics_values.append(metric_value)
                    if metric_key == stateful_metric:
                        stateful_metric_values.append(metric_value)
                    if metric_key == transactional_metric:
                        transactional_metric_values.append(metric_value)

        get_keys(agent_hostname)

        assert all(x > y for x, y in zip(persistent_metrics_values, persistent_metrics_values[1:]))
        assert all(x > y for x, y in zip(stateful_metric_values, stateful_metric_values[1:]))
        assert all(x > y for x, y in zip(transactional_metric_values, transactional_metric_values[1:]))

    util.wait_until(wait_for_metrics, 60, 3)
