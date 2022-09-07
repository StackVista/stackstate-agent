import util
import json
from util import assert_metrics, match_partial_event
from ststest import TopicTopologyMatcher

testinfra_hosts = ["local"]


def test_agent_sample_integration_generic_events(cliv1):

    def wait_for_events():
        json_data = cliv1.topic_api("sts_generic_events")
        with open("./topic-agent-integration-sample-sts-generic-events.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        service_event = {
            "name": "service-check.service-check",
            "title": "stackstate.agent.check_status",
            "eventType": "service-check",
            "tags": {
                "source_type_name": "service-check",
                "status": "OK",
                "check": "cpu"
            },
        }
        assert match_partial_event(service_event, json_data), f"no matches found for Kubernetes event: {service_event}"

        http_event = {
            "name": "HTTP_TIMEOUT",
            "title": "URL timeout",
            "eventType": "HTTP_TIMEOUT",
            "tags": {
                "source_type_name": "HTTP_TIMEOUT"
            },
            "message": "Http request to http://localhost timed out after 5.0 seconds."
        }
        assert match_partial_event(http_event, json_data), f"no matches found for Kubernetes event: {service_event}"

    util.wait_until(wait_for_events, 60, 5)


def test_agent_integration_sample_metrics(host, cliv1):
    expected = {'system.cpu.usage', 'location.availability', '2xx.responses', '5xx.responses', 'check_runs'}
    json_data = cliv1.topic_api("sts_multi_metrics")

    with open("./topic-agent-integration-sample-sts-metrics.json", 'w') as f:
            json.dump(json_data, f, indent=4)

    assert_metrics(host, json_data, expected)


def test_agent_integration_sample_topology_events(host, cliv1):

    def wait_for_topology_events():
        json_data = cliv1.topic_api("sts_topology_events")
        with open("./topic-agent-integration-sample-sts-topology-events.json", 'w') as f:
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
        ) is not None

    util.wait_until(wait_for_topology_events, 60, 3)


def test_agent_integration_sample_health_synchronization(host, cliv1):

    def wait_for_health_messages():
        json_data = cliv1.topic_api("sts_intake_health", 100)
        with open("./topic-agent-integration-sample-sts-health-messages.json", 'w') as f:
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
        assert _health_contains_payload(
            {
                "IntakeHealthCheckStates": {
                    "consistencyModel": "REPEAT_SNAPSHOTS",
                    "intakeCheckStates": [
                        {"data": "{\"checkStateId\":\"id\",\"health\":\"CRITICAL\",\"message\":\"msg\",\"name\":\"name\",\"topologyElementIdentifier\":\"identifier\"}"}
                    ]
                }
            }
        ) is not None

    util.wait_until(wait_for_health_messages, 60, 3)


def test_agent_integration_sample_topology_topic_api(host, agent_hostname, cliv1):

    agent_integration_sample_topology = TopicTopologyMatcher()\
        .component("this-host-assertion", name=r"this-host", domain=r"Webshop")\
        .component("some-application-assertion", name=r"some-application")\
        .component("delete-test-host-assertion", name=r"delete-test-host")\
        .delete(r"urn:example:/host:host_for_deletion")

    def assert_topology():
        topology_result = cliv1.topology_topic(topic="sts_topo_agent-integration_sample", limit=20)

        match_result = agent_integration_sample_topology.find(topology_result)
        match_result.assert_exact_match(strict=False)

    util.wait_until(assert_topology, 60, 3)


def test_agent_integration_monitoring_topology_topic_api(host, agent_hostname, cliv1):
    agent_integration_sample_topology = TopicTopologyMatcher() \
        .component("stackstate-agent-assertion", name=rf"StackState Agent:{agent_hostname}") \
        .component("agent-integration-assertion", name=rf"{agent_hostname}:agent-integration") \
        .component("agent-integration-sample-assertion", name=r"agent-integration:sample")

    def assert_topology():
        topology_result = cliv1.topology_topic(topic="sts_topo_agent_integrations", limit=20)

        match_result = agent_integration_sample_topology.find(topology_result)
        match_result.assert_exact_match(strict=False)

    util.wait_until(assert_topology, 60, 3)


# def test_agent_integration_sample_topology(host, agent_hostname, cliv1):

#     def assert_topology():
#         json_data = cliv1.topic_api("sts_topo_agent_integrations", 1500)

#         with open("./topic-agent-integration-sample-sts-topo-agent-integrations.json", 'w') as f:
#             json.dump(json_data, f, indent=4)

#         components = [
#             {
#                 "assertion": "Should find the this-host component",
#                 "type": "Host",
#                 "external_id": lambda e_id: "urn:example:/host:this_host" == e_id,
#                 "data": lambda d: d == {
#                     "checks": [
#                         {
#                             "critical_value": 90,
#                             "deviating_value": 75,
#                             "is_metric_maximum_average_check": 1,
#                             "max_window": 300000,
#                             "name": "Max CPU Usage (Average)",
#                             "remediation_hint": "There is too much activity on this host",
#                             "stream_id": -1
#                         },
#                         {
#                             "critical_value": 90,
#                             "deviating_value": 75,
#                             "is_metric_maximum_last_check": 1,
#                             "max_window": 300000,
#                             "name": "Max CPU Usage (Last)",
#                             "remediation_hint": "There is too much activity on this host",
#                             "stream_id": -1
#                         },
#                         {
#                             "critical_value": 5,
#                             "deviating_value": 10,
#                             "is_metric_minimum_average_check": 1,
#                             "max_window": 300000,
#                             "name": "Min CPU Usage (Average)",
#                             "remediation_hint": "There is too few activity on this host",
#                             "stream_id": -1
#                         },
#                         {
#                             "critical_value": 5,
#                             "deviating_value": 10,
#                             "is_metric_minimum_last_check": 1,
#                             "max_window": 300000,
#                             "name": "Min CPU Usage (Last)",
#                             "remediation_hint": "There is too few activity on this host",
#                             "stream_id": -1
#                         }
#                     ],
#                     "domain": "Webshop",
#                     "environment": "Production",
#                     "identifiers": [
#                         "another_identifier_for_this_host"
#                     ],
#                     "labels": [
#                         "host:this_host",
#                         "region:eu-west-1"
#                     ],
#                     "layer": "Machines",
#                     "metrics": [
#                         {
#                             "aggregation": "MEAN",
#                             "conditions": [
#                                 {
#                                     "key": "tags.hostname",
#                                     "value": "this-host"
#                                 },
#                                 {
#                                     "key": "tags.region",
#                                     "value": "eu-west-1"
#                                 }
#                             ],
#                             "metric_field": "system.cpu.usage",
#                             "name": "Host CPU Usage",
#                             "priority": "HIGH",
#                             "stream_id": -1,
#                             "unit_of_measure": "Percentage"
#                         },
#                         {
#                             "aggregation": "MEAN",
#                             "conditions": [
#                                 {
#                                     "key": "tags.hostname",
#                                     "value": "this-host"
#                                 },
#                                 {
#                                     "key": "tags.region",
#                                     "value": "eu-west-1"
#                                 }
#                             ],
#                             "metric_field": "location.availability",
#                             "name": "Host Availability",
#                             "priority": "HIGH",
#                             "stream_id": -2,
#                             "unit_of_measure": "Percentage"
#                         }
#                     ],
#                     "name": "this-host",
#                     "tags": [
#                         "integration-type:agent-integration",
#                         "integration-url:sample"
#                     ]
#                 }
#             },
#             {
#                 "assertion": "Should find the some-application component",
#                 "type": "Application",
#                 "external_id": lambda e_id: "urn:example:/application:some_application" == e_id,
#                 "data": lambda d: d == {
#                     "checks": [
#                         {
#                             "critical_value": 75,
#                             "denominator_stream_id": -1,
#                             "deviating_value": 50,
#                             "is_metric_maximum_ratio_check": 1,
#                             "max_window": 300000,
#                             "name": "OK vs Error Responses (Maximum)",
#                             "numerator_stream_id": -2
#                         },
#                         {
#                             "critical_value": 70,
#                             "deviating_value": 50,
#                             "is_metric_maximum_percentile_check": 1,
#                             "max_window": 300000,
#                             "name": "Error Response 99th Percentile",
#                             "percentile": 99,
#                             "stream_id": -2
#                         },
#                         {
#                             "critical_value": 75,
#                             "denominator_stream_id": -1,
#                             "deviating_value": 50,
#                             "is_metric_failed_ratio_check": 1,
#                             "max_window": 300000,
#                             "name": "OK vs Error Responses (Failed)",
#                             "numerator_stream_id": -2
#                         },
#                         {
#                             "critical_value": 5,
#                             "deviating_value": 10,
#                             "is_metric_minimum_percentile_check": 1,
#                             "max_window": 300000,
#                             "name": "Success Response 99th Percentile",
#                             "percentile": 99,
#                             "stream_id": -1
#                         }
#                     ],
#                     "domain": "Webshop",
#                     "environment": "Production",
#                     "identifiers": [
#                         "another_identifier_for_some_application"
#                     ],
#                     "labels": [
#                         "application:some_application",
#                         "region:eu-west-1",
#                         "hosted_on:this-host"
#                     ],
#                     "layer": "Applications",
#                     "metrics": [
#                         {
#                             "aggregation": "MEAN",
#                             "conditions": [
#                                 {
#                                     "key": "tags.application",
#                                     "value": "some_application"
#                                 },
#                                 {
#                                     "key": "tags.region",
#                                     "value": "eu-west-1"
#                                 }
#                             ],
#                             "metric_field": "2xx.responses",
#                             "name": "2xx Responses",
#                             "priority": "HIGH",
#                             "stream_id": -1,
#                             "unit_of_measure": "Count"
#                         },
#                         {
#                             "aggregation": "MEAN",
#                             "conditions": [
#                                 {
#                                     "key": "tags.application",
#                                     "value": "some_application"
#                                 },
#                                 {
#                                     "key": "tags.region",
#                                     "value": "eu-west-1"
#                                 }
#                             ],
#                             "metric_field": "5xx.responses",
#                             "name": "5xx Responses",
#                             "priority": "HIGH",
#                             "stream_id": -2,
#                             "unit_of_measure": "Count"
#                         }
#                     ],
#                     "name": "some-application",
#                     "tags": [
#                         "integration-type:agent-integration",
#                         "integration-url:sample"
#                     ],
#                     "version": "0.2.0"
#                 }
#             },
#             {
#                 "assertion": "Should find the stackstate-agent component",
#                 "type": "stackstate-agent",
#                 "external_id": lambda e_id: (f"urn:stackstate-agent:/{agent_hostname}") == e_id,
#                 "data": lambda d: d == {
#                     "hostname": agent_hostname,
#                     "identifiers": [
#                         f"urn:process:/%s:%s" % (agent_hostname, d["identifiers"][0][len("urn:process:/%s:" % agent_hostname):])
#                     ],
#                     "name": f"StackState Agent:{agent_hostname}",
#                     "tags": [
#                         f"hostname:{agent_hostname}",
#                         "stackstate-agent"
#                     ]
#                 }
#             },
#             {
#                 "assertion": "Should find the agent-integration integration component",
#                 "type": "agent-integration",
#                 "external_id": lambda e_id: ("urn:agent-integration:/%s:agent-integration" % agent_hostname) == e_id,
#                 "data": lambda d: d == {
#                     "checks": [
#                         {
#                             "is_service_check_health_check": 1,
#                             "name": "Integration Health",
#                             "stream_id": -1
#                         }
#                     ],
#                     "hostname": agent_hostname,
#                     "integration": "agent-integration",
#                     "name": "%s:agent-integration" % agent_hostname,
#                     "service_checks": [
#                         {
#                             "conditions": [
#                                 {
#                                     "key": "host",
#                                     "value": agent_hostname
#                                 },
#                                 {
#                                     "key": "tags.integration-type",
#                                     "value": "agent-integration"
#                                 }
#                             ],
#                             "name": "Service Checks",
#                             "stream_id": -1
#                         }
#                     ],
#                     "tags": [
#                         "hostname:%s" % agent_hostname,
#                         "integration-type:agent-integration"
#                     ]
#                 }
#             },
#             {
#                 "assertion": "Should find the agent-integration-instance component",
#                 "type": "agent-integration-instance",
#                 "external_id": lambda e_id: ("urn:agent-integration-instance:/%s:agent-integration:sample" % agent_hostname) == e_id,
#                 "data": lambda d: d == {
#                     "checks": [
#                         {
#                             "is_service_check_health_check": 1,
#                             "name": "Integration Instance Health",
#                             "stream_id": -1
#                         }
#                     ],
#                     "hostname": agent_hostname,
#                     "integration": "agent-integration",
#                     "name": "agent-integration:sample",
#                     "service_checks": [
#                         {
#                             "conditions": [
#                                 {
#                                     "key": "host",
#                                     "value": agent_hostname
#                                 },
#                                 {
#                                     "key": "tags.integration-type",
#                                     "value": "agent-integration"
#                                 },
#                                 {
#                                     "key": "tags.integration-url",
#                                     "value": "sample"
#                                 }
#                             ],
#                             "name": "Service Checks",
#                             "stream_id": -1
#                         }
#                     ],
#                     "tags": [
#                         "hostname:%s" % agent_hostname,
#                         "integration-type:agent-integration",
#                         "integration-url:sample"
#                     ]
#                 }
#             }
#         ]

#         for c in components:
#             print("Running assertion for: " + c["assertion"])
#             assert util.component_data(
#                 json_data=json_data,
#                 type_name=c["type"],
#                 external_id_assert_fn=c["external_id"],
#                 data_assert_fn=c["data"],
#             ) is not None

#         assert util.delete_topo_element_data(json_data, "urn:example:/host:host_for_deletion")

#     util.wait_until(assert_topology, 10, 3)
