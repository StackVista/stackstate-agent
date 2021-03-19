import json
import os

import util

from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('kubernetes-cluster-agent')


def _get_key_value(tag_list):
    for key, value in (pair.split(':', 1) for pair in tag_list):
        yield key, value


def kubernetes_event_data(event, json_data):
    for message in json_data["messages"]:
        p = message["message"]
        if "GenericEvent" in p:
            _data = p["GenericEvent"]
            if _data == dict(_data, **event):
                return _data
    return None


def test_agent_integration_sample_metrics(host):
    hostname = host.ansible.get_variables()["inventory_hostname"]
    url = "http://localhost:7070/api/topic/sts_multi_metrics?limit=1000"

    def wait_for_metrics():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-agent-integration-sample-sts-multi-metrics.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        def get_keys(m_host):
            return set(
                ''.join(message["message"]["MultiMetric"]["values"].keys())
                for message in json_data["messages"]
                if message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                message["message"]["MultiMetric"]["host"] == m_host
            )

        expected = {'system.cpu.usage', 'location.availability', '2xx.responses', '5xx.responses', 'check_runs'}
        assert all([expectedMetric for expectedMetric in expected if expectedMetric in get_keys(hostname)])

    util.wait_until(wait_for_metrics, 180, 3)


def test_agent_integration_sample_topology(host):
    def assert_topology():
        topo_url = "http://localhost:7070/api/topic/sts_topo_agent_integrations?limit=1500"
        data = host.check_output('curl "{}"'.format(topo_url))
        json_data = json.loads(data)
        with open("./topic-agent-integration-sample-topo-agent-integrations.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        components = [
            {
                "assertion": "Should find the this-host component",
                "type": "Host",
                "external_id": lambda e_id: "urn:example:/host:this_host" == e_id,
                "data": lambda d: d == {
                    "checks": [
                        {
                            "critical_value": 90,
                            "deviating_value": 75,
                            "is_metric_maximum_average_check": 1,
                            "max_window": 300000,
                            "name": "Max CPU Usage (Average)",
                            "remediation_hint": "There is too much activity on this host",
                            "stream_id": -1
                        },
                        {
                            "critical_value": 90,
                            "deviating_value": 75,
                            "is_metric_maximum_last_check": 1,
                            "max_window": 300000,
                            "name": "Max CPU Usage (Last)",
                            "remediation_hint": "There is too much activity on this host",
                            "stream_id": -1
                        },
                        {
                            "critical_value": 5,
                            "deviating_value": 10,
                            "is_metric_minimum_average_check": 1,
                            "max_window": 300000,
                            "name": "Min CPU Usage (Average)",
                            "remediation_hint": "There is too few activity on this host",
                            "stream_id": -1
                        },
                        {
                            "critical_value": 5,
                            "deviating_value": 10,
                            "is_metric_minimum_last_check": 1,
                            "max_window": 300000,
                            "name": "Min CPU Usage (Last)",
                            "remediation_hint": "There is too few activity on this host",
                            "stream_id": -1
                        }
                    ],
                    "domain": "Webshop",
                    "environment": "Production",
                    "identifiers": [
                        "another_identifier_for_this_host"
                    ],
                    "labels": [
                        "host:this_host",
                        "region:eu-west-1"
                    ],
                    "layer": "Machines",
                    "metrics": [
                        {
                            "aggregation": "MEAN",
                            "conditions": [
                                {
                                    "key": "tags.hostname",
                                    "value": "this-host"
                                },
                                {
                                    "key": "tags.region",
                                    "value": "eu-west-1"
                                }
                            ],
                            "metric_field": "system.cpu.usage",
                            "name": "Host CPU Usage",
                            "priority": "HIGH",
                            "stream_id": -1,
                            "unit_of_measure": "Percentage"
                        },
                        {
                            "aggregation": "MEAN",
                            "conditions": [
                                {
                                    "key": "tags.hostname",
                                    "value": "this-host"
                                },
                                {
                                    "key": "tags.region",
                                    "value": "eu-west-1"
                                }
                            ],
                            "metric_field": "location.availability",
                            "name": "Host Availability",
                            "priority": "HIGH",
                            "stream_id": -2,
                            "unit_of_measure": "Percentage"
                        }
                    ],
                    "name": "this-host",
                    "tags": [
                        "integration-type:agent-integration",
                        "integration-url:sample"
                    ]
                }
            },
            {
                "assertion": "Should find the some-application component",
                "type": "Application",
                "external_id": lambda e_id: "urn:example:/application:some_application" == e_id,
                "data": lambda d: d == {
                    "checks": [
                        {
                            "critical_value": 75,
                            "denominator_stream_id": -1,
                            "deviating_value": 50,
                            "is_metric_maximum_ratio_check": 1,
                            "max_window": 300000,
                            "name": "OK vs Error Responses (Maximum)",
                            "numerator_stream_id": -2
                        },
                        {
                            "critical_value": 70,
                            "deviating_value": 50,
                            "is_metric_maximum_percentile_check": 1,
                            "max_window": 300000,
                            "name": "Error Response 99th Percentile",
                            "percentile": 99,
                            "stream_id": -2
                        },
                        {
                            "critical_value": 75,
                            "denominator_stream_id": -1,
                            "deviating_value": 50,
                            "is_metric_failed_ratio_check": 1,
                            "max_window": 300000,
                            "name": "OK vs Error Responses (Failed)",
                            "numerator_stream_id": -2
                        },
                        {
                            "critical_value": 5,
                            "deviating_value": 10,
                            "is_metric_minimum_percentile_check": 1,
                            "max_window": 300000,
                            "name": "Success Response 99th Percentile",
                            "percentile": 99,
                            "stream_id": -1
                        }
                    ],
                    "domain": "Webshop",
                    "environment": "Production",
                    "identifiers": [
                        "another_identifier_for_some_application"
                    ],
                    "labels": [
                        "application:some_application",
                        "region:eu-west-1",
                        "hosted_on:this-host"
                    ],
                    "layer": "Applications",
                    "metrics": [
                        {
                            "aggregation": "MEAN",
                            "conditions": [
                                {
                                    "key": "tags.application",
                                    "value": "some_application"
                                },
                                {
                                    "key": "tags.region",
                                    "value": "eu-west-1"
                                }
                            ],
                            "metric_field": "2xx.responses",
                            "name": "2xx Responses",
                            "priority": "HIGH",
                            "stream_id": -1,
                            "unit_of_measure": "Count"
                        },
                        {
                            "aggregation": "MEAN",
                            "conditions": [
                                {
                                    "key": "tags.application",
                                    "value": "some_application"
                                },
                                {
                                    "key": "tags.region",
                                    "value": "eu-west-1"
                                }
                            ],
                            "metric_field": "5xx.responses",
                            "name": "5xx Responses",
                            "priority": "HIGH",
                            "stream_id": -2,
                            "unit_of_measure": "Count"
                        }
                    ],
                    "name": "some-application",
                    "tags": [
                        "integration-type:agent-integration",
                        "integration-url:sample"
                    ],
                    "version": "0.2.0"
                }
            }
        ]

        for c in components:
            print("Running assertion for: " + c["assertion"])
            assert util.component_data(
                json_data=json_data,
                type_name=c["type"],
                external_id_assert_fn=c["external_id"],
                data_assert_fn=c["data"],
            ) is not None

    util.wait_until(assert_topology, 30, 3)


def test_agent_integration_sample_events(host):
    url = "http://localhost:7070/api/topic/sts_generic_events?limit=1000"

    def wait_for_events():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
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
        assert kubernetes_event_data(service_event, json_data) is not None

        http_event = {
            "name": "HTTP_TIMEOUT",
            "title": "URL timeout",
            "eventType": "HTTP_TIMEOUT",
            "tags": {
                "source_type_name": "HTTP_TIMEOUT"
            },
            "message": "Http request to http://localhost timed out after 5.0 seconds."
        }
        assert kubernetes_event_data(http_event, json_data) is not None

    util.wait_until(wait_for_events, 60, 3)


def test_agent_integration_sample_topology_events(host):
    url = "http://localhost:7070/api/topic/sts_topology_events?limit=1000"

    def wait_for_topology_events():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
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
