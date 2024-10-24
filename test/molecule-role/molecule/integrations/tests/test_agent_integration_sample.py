import json
import os

from testinfra.utils.ansible_runner import AnsibleRunner

import util

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('agent-integrations')


def _get_key_value(tag_list):
    for key, value in (pair.split(':', 1) for pair in tag_list):
        yield key, value


def test_agent_integration_sample_metrics(host, hostname):
    url = "http://localhost:7070/api/topic/sts_multi_metrics?limit=1000"

    def wait_for_metrics():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-agent-integration-sample-sts-multi-metrics.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        def get_keys(m_host):
            data_set = set()
            for message in json_data["messages"]:
                multi_metric = message["message"]["MultiMetric"]
                if multi_metric["name"] == "convertedMetric" and multi_metric["host"] == m_host:
                    for key in multi_metric["values"].keys():
                        data_set.add(key)
            return data_set

        keys = get_keys(hostname)

        expected = {'system.cpu.usage', 'location.availability', '2xx.responses', '5xx.responses', 'raw.metrics'}
        sets_difference = expected.difference(keys)
        if len(sets_difference) > 0:
            print("Missing keys inside the topic response:", (expected.difference(keys)))
        assert expected.issubset(keys) is True

        expected.add('invalid.value.does.not.exist')
        assert expected.issubset(keys) is False

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_integration_sample_topology(host, hostname):
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
            },
            {
                "assertion": "Should find the stackstate-agent component",
                "type": "stackstate-agent",
                "external_id": lambda e_id: ("urn:stackstate-agent:/%s" % hostname) == e_id,
                "data": lambda d: d == {
                    "hostname": "agent-integrations",
                    "identifiers": [
                        "urn:process:/%s:%s" % (hostname, d["identifiers"][0][len("urn:process:/%s:" % hostname):])
                    ],
                    "name": "StackState Agent:agent-integrations",
                    "tags": [
                        "hostname:agent-integrations",
                        "stackstate-agent"
                    ]
                }
            },
            {
                "assertion": "Should find the agent-integration integration component",
                "type": "agent-integration",
                "external_id": lambda e_id: ("urn:agent-integration:/%s:agent-integration" % hostname) == e_id,
                "data": lambda d: d == {
                    "checks": [
                        {
                            "is_service_check_health_check": 1,
                            "name": "Integration Health",
                            "stream_id": -1
                        }
                    ],
                    "hostname": hostname,
                    "integration": "agent-integration",
                    "name": "%s:agent-integration" % hostname,
                    "service_checks": [
                        {
                            "conditions": [
                                {
                                    "key": "host",
                                    "value": hostname
                                },
                                {
                                    "key": "tags.integration-type",
                                    "value": "agent-integration"
                                }
                            ],
                            "name": "Service Checks",
                            "stream_id": -1
                        }
                    ],
                    "tags": [
                        "hostname:%s" % hostname,
                        "integration-type:agent-integration"
                    ]
                }
            },
            {
                "assertion": "Should find the agent-integration-instance component",
                "type": "agent-integration-instance",
                "external_id": lambda e_id: ("urn:agent-integration-instance:/%s:agent-integration:sample" % hostname) == e_id,
                "data": lambda d: d == {
                    "checks": [
                        {
                            "is_service_check_health_check": 1,
                            "name": "Integration Instance Health",
                            "stream_id": -1
                        }
                    ],
                    "hostname": hostname,
                    "integration": "agent-integration",
                    "name": "agent-integration:sample",
                    "service_checks": [
                        {
                            "conditions": [
                                {
                                    "key": "host",
                                    "value": hostname
                                },
                                {
                                    "key": "tags.integration-type",
                                    "value": "agent-integration"
                                },
                                {
                                    "key": "tags.integration-url",
                                    "value": "sample"
                                }
                            ],
                            "name": "Service Checks",
                            "stream_id": -1
                        }
                    ],
                    "tags": [
                        "hostname:%s" % hostname,
                        "integration-type:agent-integration",
                        "integration-url:sample"
                    ]
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

        assert util.delete_topo_element_data(json_data, "urn:example:/host:host_for_deletion")

    util.wait_until(assert_topology, 30, 3)


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


def test_agent_integration_sample_health_synchronization(host):
    url = "http://localhost:7070/api/topic/sts_intake_health?limit=100"

    def wait_for_health_messages():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
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
                "repeatIntervalMs": 30000
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


def test_agent_integration_sample_persistent_cache(host):
    def wait_for_persistent_cache():
        data = host.check_output("docker exec ubuntu_stackstate-agent_1 "
                                 "sh -c 'ls -d /opt/stackstate-agent/run/agent_integration_sample/* | xargs cat'")
        json_data = json.loads(data)
        with open("./persistent-cache-agent-integration-sample.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        sample_data = {
            "name": "some-application",
            "domain": "Webshop",
            "layer": "Applications",
            "identifiers": ["another_identifier_for_some_application"],
            "labels": ["application:some_application", "region:eu-west-1", "hosted_on:this-host"],
            "environment": "Production",
            "version": "0.2.0"
        }

        assert json_data == sample_data

    util.wait_until(wait_for_persistent_cache, 60, 3)
