import logging
import time
from typing import Optional

import pytest

from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from splunk_testing_base import SplunkBase, SplunkEvent
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


@pytest.mark.order(1)
def test_splunk_event(agent: AgentTestingBase,
                      splunk: SplunkBase,
                      cliv1: CLIv1,
                      simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    event: SplunkEvent = splunk.event.publish_event()

    logging.info(f"Looking for the event host value: {event.get('host')}")

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_generic_events",
                                    query="message.GenericEvent.tags",
                                    contains_dict={
                                        "source_type_name": "generic_splunk_event",
                                        "host": event.get("host"),
                                        "description": event.get("description"),
                                        "status": event.get("status")
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: simulator_dump())

    logging.info(f"Found the following results: {result}")


@pytest.mark.order(2)
def test_splunk_multiple_events(agent: AgentTestingBase,
                                splunk: SplunkBase,
                                simulator_dump,
                                cliv1: CLIv1):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    event_a: SplunkEvent = splunk.event.publish_event()

    time.sleep(30)

    event_b: SplunkEvent = splunk.event.publish_event()

    logging.info(f"Looking for the event host value: {event_a.get('host')}")
    logging.info(f"Looking for the event host value: {event_b.get('host')}")

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_generic_events",
                                    query="message.GenericEvent.tags",
                                    contains_dict={
                                        "source_type_name": "generic_splunk_event",
                                        "host": event_a.get("host"),
                                        "description": event_a.get("description"),
                                        "status": event_a.get("status")
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5)

    logging.info(f"Found the following results: {result}")

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_generic_events",
                                    query="message.GenericEvent.tags",
                                    contains_dict={
                                        "source_type_name": "generic_splunk_event",
                                        "host": event_b.get("host"),
                                        "description": event_b.get("description"),
                                        "status": event_b.get("status")
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: simulator_dump())

    logging.info(f"Found the following results: {result}")


# Stateful State
# We will publish health while the agent is active and wait for it
# When we find it then we will stop the agent and post a second health state
# After a few minutes we start the agent up again
# And wait to find the second health state
@pytest.mark.order(3)
def test_splunk_event_stateful_state(agent: AgentTestingBase,
                                     cliv1: CLIv1,
                                     splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_event(expect_failure: bool = False,
                   expected_event: SplunkEvent = None) -> Optional[SplunkEvent]:
        event: SplunkEvent = expected_event

        if event is None:
            event = splunk.event.publish_event()

        try:
            # Wait until we find the results in the Topic
            result = wait_until_topic_match(cliv1,
                                            topic="sts_generic_events",
                                            query="message.GenericEvent.tags",
                                            contains_dict={
                                                "source_type_name": "generic_splunk_event",
                                                "host": event.get("host"),
                                                "description": event.get("description"),
                                                "status": event.get("status")
                                            },
                                            first_match=True,
                                            timeout=180,
                                            period=5)

            logging.info(f"Found the following results: {result}")

        except Exception as e:
            if expect_failure is True:
                return event
            else:
                raise e

        if expect_failure is True:
            raise Exception("Metric should not exist but did not fail with a exception")
        else:
            return event

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    event_posted_while_agent_was_down: Optional[SplunkEvent] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_event_while_agent_is_stopped():
        nonlocal event_posted_while_agent_was_down
        event_posted_while_agent_was_down = post_event(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_event_after_agent_started():
        post_event(expected_event=event_posted_while_agent_was_down)

    # Run a stateful test for the agent
    agent.stateful_state_run_cycle_test(
        func_before_agent_stop=post_event,
        func_after_agent_stop=find_event_while_agent_is_stopped,
        func_after_agent_startup=find_event_after_agent_started
    )


# Transactional State
# We will produce metric while the routing is open
# Then we will close the routes, post another metric state and make sure that the metric does not exist
# After that we will open the routes and test if the metric eventually end up in STS
@pytest.mark.order(4)
def test_splunk_event_transactional_check(agent: AgentTestingBase,
                                          cliv1: CLIv1,
                                          splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_event(expect_failure: bool = False,
                   expected_event: SplunkEvent = None,
                   timeout: int = 180) -> Optional[SplunkEvent]:
        event: SplunkEvent = expected_event

        if event is None:
            event = splunk.event.publish_event()

        try:
            # Wait until we find the results in the Topic
            result = wait_until_topic_match(cliv1,
                                            topic="sts_generic_events",
                                            query="message.GenericEvent.tags",
                                            contains_dict={
                                                "source_type_name": "generic_splunk_event",
                                                "host": event.get("host"),
                                                "description": event.get("description"),
                                                "status": event.get("status")
                                            },
                                            first_match=True,
                                            timeout=timeout,
                                            period=5)

            logging.info(f"Found the following results: {result}")

        except Exception as e:
            if expect_failure is True:
                return event
            else:
                raise e

        if expect_failure is True:
            raise Exception("Metric should not exist but did not fail with a exception")
        else:
            return event

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    event_posted_while_agent_was_down: Optional[SplunkEvent] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_event_while_routes_is_blocked():
        nonlocal event_posted_while_agent_was_down
        event_posted_while_agent_was_down = post_event(expect_failure=True, timeout=180)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_event_while_routes_is_open():
        post_event(expected_event=event_posted_while_agent_was_down)

    # Run a stateful test for the agent
    agent.transactional_run_cycle_test(
        func_before_blocking_routes=post_event,
        func_after_blocking_routes=find_event_while_routes_is_blocked,
        rerun_func_unblocking_blocked_routes=find_event_while_routes_is_open
    )
