import logging
import time
from typing import Optional

import pytest

from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from splunk_testing_base import SplunkBase, SplunkMetric
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


@pytest.mark.order(1)
def test_splunk_metrics(agent: AgentTestingBase,
                        splunk: SplunkBase,
                        cliv1: CLIv1,
                        simulator_dump,
                        request):
    # Get the current machine time on the agent host machine
    start_date_time = agent.get_current_time_on_agent_machine()

    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    metric: SplunkMetric = splunk.metric.publish_metric()

    logging.info(f"Looking for the raw.metric value: {metric.get('value')}")

    def failure_dump():
        simulator_dump()
        agent.dump_logs(request, start_date_time)

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_multi_metrics",
                                    query="message.MultiMetric.values",
                                    contains_dict={
                                        "raw.metrics": int(metric.get("value")),
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: failure_dump())

    logging.info(f"Found the following results: {result}")


@pytest.mark.order(2)
def test_splunk_multiple_metrics(agent: AgentTestingBase,
                                 splunk: SplunkBase,
                                 simulator_dump,
                                 request,
                                 cliv1: CLIv1):
    # Get the current machine time on the agent host machine
    start_date_time = agent.get_current_time_on_agent_machine()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    metric_a: SplunkMetric = splunk.metric.publish_metric()

    # Add a delay between data or they will be seen as the same data
    time.sleep(30)

    metric_b: SplunkMetric = splunk.metric.publish_metric()

    logging.info(f"Looking for the raw.metric A value: {metric_a.get('value')}")
    logging.info(f"Looking for the raw.metric B value: {metric_b.get('value')}")

    def failure_dump():
        simulator_dump()
        agent.dump_logs(request, start_date_time)

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_multi_metrics",
                                    query="message.MultiMetric.values",
                                    contains_dict={
                                        "raw.metrics": int(metric_a.get("value")),
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5)

    logging.info(f"Found the following results: {result}")

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_multi_metrics",
                                    query="message.MultiMetric.values",
                                    contains_dict={
                                        "raw.metrics": int(metric_b.get("value")),
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: failure_dump())

    logging.info(f"Found the following results: {result}")


# Stateful State
# We will publish health while the agent is active and wait for it
# When we find it then we will stop the agent and post a second health state
# After a few minutes we start the agent up again
# And wait to find the second health state
@pytest.mark.order(3)
def test_splunk_metric_stateful_state(agent: AgentTestingBase,
                                      cliv1: CLIv1,
                                      splunk: SplunkBase,
                                      simulator_dump,
                                      request):
    # Get the current machine time on the agent host machine
    start_date_time = agent.get_current_time_on_agent_machine()

    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_metric(expect_failure: bool = False,
                    expected_metric: SplunkMetric = None) -> Optional[SplunkMetric]:
        # Add sleep to make sure the splunk data has time between data points and when the agent started up
        time.sleep(30)

        metric: SplunkMetric = expected_metric

        if metric is None:
            metric = splunk.metric.publish_metric()

        try:
            # Wait until we find the results in the Topic
            result = wait_until_topic_match(cliv1,
                                            topic="sts_multi_metrics",
                                            query="message.MultiMetric.values",
                                            contains_dict={
                                                "raw.metrics": int(metric.get("value")),
                                            },
                                            first_match=True,
                                            timeout=180,
                                            period=5)

            logging.info(f"Found the following results: {result}")

        except Exception as e:
            if expect_failure is True:
                return metric
            else:
                raise e

        if expect_failure is True:
            raise Exception("Metric should not exist but did not fail with a exception")
        else:
            return metric

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    metric_posted_while_agent_was_down: Optional[SplunkMetric] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_metric_while_agent_is_stopped():
        nonlocal metric_posted_while_agent_was_down
        metric_posted_while_agent_was_down = post_metric(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_metric_after_agent_started():
        post_metric(expected_metric=metric_posted_while_agent_was_down)

        # Post extra health to make sure new ones also work
        post_metric()

    try:
        # Run a stateful test for the agent
        agent.stateful_state_run_cycle_test(
            func_before_agent_stop=post_metric,
            func_after_agent_stop=find_metric_while_agent_is_stopped,
            func_after_agent_startup=find_metric_after_agent_started
        )

    except Exception as e:
        simulator_dump()
        agent.dump_logs(request, start_date_time)
        raise e


# Transactional State
# We will produce metric while the routing is open
# Then we will close the routes, post another metric state and make sure that the metric does not exist
# After that we will open the routes and test if the metric eventually end up in STS
@pytest.mark.order(4)
def test_splunk_metric_transactional_check(agent: AgentTestingBase,
                                           cliv1: CLIv1,
                                           splunk: SplunkBase,
                                           simulator_dump,
                                           request):
    # Get the current machine time on the agent host machine
    start_date_time = agent.get_current_time_on_agent_machine()

    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_metric(expect_failure: bool = False,
                    expected_metric: SplunkMetric = None,
                    timeout: int = 180) -> Optional[SplunkMetric]:
        metric: SplunkMetric = expected_metric

        if metric is None:
            metric = splunk.metric.publish_metric()

        try:
            # Wait until we find the results in the Topic
            result = wait_until_topic_match(cliv1,
                                            topic="sts_multi_metrics",
                                            query="message.MultiMetric.values",
                                            contains_dict={
                                                "raw.metrics": int(metric.get("value")),
                                            },
                                            first_match=True,
                                            timeout=timeout,
                                            period=5)

            logging.info(f"Found the following results: {result}")

        except Exception as e:
            if expect_failure is True:
                return metric
            else:
                raise e

        if expect_failure is True:
            raise Exception("Metric should not exist but did not fail with a exception")
        else:
            return metric

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    metric_posted_while_agent_was_down: Optional[SplunkMetric] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_metric_while_routes_is_blocked():
        nonlocal metric_posted_while_agent_was_down
        logging.info(f"Posting Metric: {metric_posted_while_agent_was_down}")
        metric_posted_while_agent_was_down = post_metric(expect_failure=True,
                                                         timeout=180)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_metric_while_routes_is_open():
        logging.info(f"Looking for Metric: {metric_posted_while_agent_was_down}")
        post_metric(expected_metric=metric_posted_while_agent_was_down,
                    timeout=180)
        # post_metric()

    try:
        # Run a stateful test for the agent
        agent.transactional_run_cycle_test(
            func_before_blocking_routes=post_metric,
            func_after_blocking_routes=find_metric_while_routes_is_blocked,
            rerun_func_unblocking_blocked_routes=find_metric_while_routes_is_open
        )

    except Exception as e:
        simulator_dump()
        agent.dump_logs(request, start_date_time)
        raise e

# sudo systemctl edit stackstate-agent.service --force --full
# Environment="LOG_PAYLOADS=true"
# Environment="STS_LOG_LEVEL=debug"
