# (C) StackState 2020
# All rights reserved
# Licensed under a 3-clause BSD style license (see LICENSE)

import json

from stackstate_checks.base import AgentCheck, StackPackInstance, MetricStream, MetricHealthChecks, \
    HealthStream, HealthStreamUrn, Health
import time
from random import seed
from random import randint

seed(1)


class AgentIntegrationSampleCheck(AgentCheck):
    def get_health_stream(self, instance):
        return HealthStream(HealthStreamUrn("integration-sample", "sample"))

    def get_instance_key(self, instance):
        return StackPackInstance("agent-integration", "sample")

    def check(self, instance):
        # gets the value of the `url` property
        instance_url = instance.get('url', 'agent-integration-sample')
        # gets the value of the `default_timeout` property or defaults to 5
        default_timeout = self.init_config.get('default_timeout', 5)
        # gets the value of the `timeout` property or defaults `default_timeout` and casts it to a float data type
        timeout = float(instance.get('timeout', default_timeout))

        this_host_availability = MetricStream("Host Availability", "location.availability",
                                              conditions={"tags.hostname": "this-host", "tags.region": "eu-west-1"},
                                              unit_of_measure="Percentage",
                                              aggregation="MEAN",
                                              priority="HIGH")

        this_host_cpu_usage = MetricStream("Host CPU Usage", "system.cpu.usage",
                                           conditions={"tags.hostname": "this-host", "tags.region": "eu-west-1"},
                                           unit_of_measure="Percentage",
                                           aggregation="MEAN",
                                           priority="HIGH")
        cpu_max_average_check = MetricHealthChecks.maximum_average(this_host_cpu_usage.identifier,
                                                                   "Max CPU Usage (Average)", 75, 90,
                                                                   remediation_hint="There is too much activity on "
                                                                                    "this host")
        cpu_max_last_check = MetricHealthChecks.maximum_last(this_host_cpu_usage.identifier,
                                                             "Max CPU Usage (Last)", 75, 90,
                                                             remediation_hint="There is too much activity on "
                                                                              "this host")
        cpu_min_average_check = MetricHealthChecks.minimum_average(this_host_cpu_usage.identifier,
                                                                   "Min CPU Usage (Average)", 10, 5,
                                                                   remediation_hint="There is too few activity on "
                                                                                    "this host")
        cpu_min_last_check = MetricHealthChecks.minimum_last(this_host_cpu_usage.identifier,
                                                             "Min CPU Usage (Last)", 10, 5,
                                                             remediation_hint="There is too few activity on this host")
        self.component("urn:example:/host:this_host", "Host",
                       data={
                           "name": "this-host",
                           "domain": "Webshop",
                           "layer": "Machines",
                           "identifiers": ["another_identifier_for_this_host"],
                           "labels": ["host:this_host", "region:eu-west-1"],
                           "environment": "Production"
                       },
                       streams=[this_host_cpu_usage, this_host_availability],
                       checks=[cpu_max_average_check, cpu_max_last_check, cpu_min_average_check, cpu_min_last_check])

        self.gauge("system.cpu.usage", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])
        self.gauge("system.cpu.usage", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])
        self.gauge("system.cpu.usage", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])
        self.gauge("location.availability", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])
        self.gauge("location.availability", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])
        self.gauge("location.availability", randint(0, 100), tags=["hostname:this-host", "region:eu-west-1"])

        some_application_2xx_responses = MetricStream("2xx Responses", "2xx.responses",
                                                      conditions={"tags.application": "some_application",
                                                                  "tags.region": "eu-west-1"},
                                                      unit_of_measure="Count",
                                                      aggregation="MEAN",
                                                      priority="HIGH")
        some_application_5xx_responses = MetricStream("5xx Responses", "5xx.responses",
                                                      conditions={"tags.application": "some_application",
                                                                  "tags.region": "eu-west-1"},
                                                      unit_of_measure="Count",
                                                      aggregation="MEAN",
                                                      priority="HIGH")
        max_response_ratio_check = MetricHealthChecks.maximum_ratio(some_application_2xx_responses.identifier,
                                                                    some_application_5xx_responses.identifier,
                                                                    "OK vs Error Responses (Maximum)",
                                                                    50, 75)
        max_percentile_response_check = MetricHealthChecks.maximum_percentile(some_application_5xx_responses.identifier,
                                                                              "Error Response 99th Percentile",
                                                                              50, 70, 99)
        failed_response_ratio_check = MetricHealthChecks.failed_ratio(some_application_2xx_responses.identifier,
                                                                      some_application_5xx_responses.identifier,
                                                                      "OK vs Error Responses (Failed)",
                                                                      50, 75)
        min_percentile_response_check = MetricHealthChecks.minimum_percentile(some_application_2xx_responses.identifier,
                                                                              "Success Response 99th Percentile",
                                                                              10, 5, 99)
        self.component("urn:example:/application:some_application", "Application",
                       data={
                           "name": "some-application",
                           "domain": "Webshop",
                           "layer": "Applications",
                           "identifiers": ["another_identifier_for_some_application"],
                           "labels": ["application:some_application", "region:eu-west-1", "hosted_on:this-host"],
                           "environment": "Production",
                           "version": "0.2.0"
                       },
                       streams=[some_application_2xx_responses, some_application_5xx_responses],
                       checks=[max_response_ratio_check, max_percentile_response_check, failed_response_ratio_check,
                               min_percentile_response_check])

        self.component("urn:example:/application:some_application/tags", "Application",
                       data={
                           "name": "some-tags",
                           "domain": "default-domain",
                           "layer": "default-layer",
                           "identifiers": ["default-identifiers"],
                           "labels": ["default:true"],
                           "environment": "default-environment",
                           "version": "0.2.0",
                           "tags": [
                               "stackstate-layer:tag-layer",
                               "stackstate-domain:tag-domain",
                               "stackstate-identifiers:tag-identifiers-a",
                               "stackstate-identifier:tag-identifier"
                           ],
                       },
                       streams=[some_application_2xx_responses, some_application_5xx_responses],
                       checks=[max_response_ratio_check, max_percentile_response_check, failed_response_ratio_check,
                               min_percentile_response_check])

        self.relation("urn:example:/application:some_application", "urn:example:/host:this_host", "IS_HOSTED_ON", {})

        self.gauge("2xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("2xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("2xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("2xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("5xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("5xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("5xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])
        self.gauge("5xx.responses", randint(0, 100), tags=["application:some_application", "region:eu-west-1"])

        self.event({
            "timestamp": int(time.time()),
            # test old api - source type name instead of event type
            "source_type_name": "HTTP_TIMEOUT",
            "msg_title": "URL timeout",
            "msg_text": "Http request to %s timed out after %s seconds." % (instance_url, timeout),
            "aggregation_key": "instance-request-%s" % instance_url
        })

        self.event({
            "timestamp": int(time.time()),
            # test old api - source type name instead of event type
            "event_type": "HTTP_TIMEOUT_EVENT_TYPE",
            "msg_title": "URL timeout (event type)",
            "msg_text": "Http request to %s timed out after %s seconds.(event type)" % (instance_url, timeout),
            "aggregation_key": "instance-request-%s" % instance_url
        })

        self.event({
            "timestamp": int(1),
            "event_type": "HTTP_TIMEOUT",
            "msg_title": "URL timeout",
            "msg_text": "Http request to %s timed out after %s seconds." % (instance_url, timeout),
            "aggregation_key": "instance-request-%s" % instance_url,
            "context": {
                "source_identifier": "source_identifier_value",
                "element_identifiers": ["urn:host:/123"],
                "source": "source_value",
                "category": "my_category",
                "data": {"big_black_hole": "here", "another_thing": 1, "test": {"1": "test"}},
                "source_links": [
                    {"title": "my_event_external_link", "url": "http://localhost"}
                ]
            }
        })

        # some logic here to test our connection and if successful:
        self.service_check("example.can_connect", AgentCheck.OK, tags=["instance_url:%s" % instance_url])

        # produce current state as a count metric for assertions
        state = instance.get('state', {'run_count': 0})
        state['run_count'] = state['run_count'] + 1
        instance.update({'state': state})
        self.count("check_runs", state['run_count'], tags=["integration:agent_integration_sample"])

        # produce health snapshot
        self.health.start_snapshot()
        self.health.check_state("id", "name", Health.CRITICAL, "identifier", "msg")
        self.health.stop_snapshot()

        # raw metrics
        self.raw("raw.metrics", hostname="hostname", value=10,
                 tags=["application:some_application", "region:eu-west-1"], timestamp=int(time.time()))
        self.raw("raw.metrics", hostname="hostname", value=10,
                 tags=["application:some_application", "region:eu-west-1"], timestamp=int(time.time()))
        self.raw("raw.metrics", value=30, tags=["no:hostname", "region:eu-west-1"], timestamp=int(time.time()))

        # delete topology element
        delete_component_id = "urn:example:/host:host_for_deletion"
        self.component(delete_component_id,  "Host",
                       data={
                            "name": "delete-test-host",
                            "domain": "Webshop",
                            "layer": "Machines",
                            "identifiers": ["another_identifier_for_delete_test_host"],
                            "labels": ["host:delete_test_host", "region:eu-west-1"],
                            "environment": "Production"
                       },)
        self.log.info("deleting " + delete_component_id)
        self.delete(delete_component_id)

        # test writing and reading from agent persistent cache
        self.log.info("Writing sample data to persistent cache")
        sample_data = {
            "name": "some-application",
            "domain": "Webshop",
            "layer": "Applications",
            "identifiers": ["another_identifier_for_some_application"],
            "labels": ["application:some_application", "region:eu-west-1", "hosted_on:this-host"],
            "environment": "Production",
            "version": "0.2.0"
        }
        self.write_persistent_cache("test_key", json.dumps(sample_data))
        self.log.info("Read key that doesn't exist in persistent cache")
        cache = self.read_persistent_cache("key_that_is_not_there")
        self.log.debug("empty cache: {} empty cache's type: {}".format(cache, type(cache)))
