import logging
import pytest

from typing import Optional
from agent_tesing_base import AgentTestingBase
from splunk_testing_base import SplunkBase, SplunkHealth
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


@pytest.mark.order(1)
def test_splunk_health(agent: AgentTestingBase,
                       splunk: SplunkBase,
                       cliv1: CLIv1,
                       simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    health: SplunkHealth = splunk.health.publish_health()

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_health_sync",
                                    query="message.HealthSyncMessage.payload.CheckStates.checkStates[*]",
                                    contains_dict={
                                        "name": health.get("name"),
                                        "checkStateId": health.get("check_state_id"),
                                        "health": health.get("health"),
                                        "message": health.get("message"),
                                        "topologyElementIdentifier": health.get("topology_element_identifier"),
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: simulator_dump())

    logging.info(f"Found the following results: {result}")


@pytest.mark.order(2)
def test_splunk_multiple_health(agent: AgentTestingBase,
                                splunk: SplunkBase,
                                cliv1: CLIv1,
                                simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    health_a: SplunkHealth = splunk.health.publish_health()
    health_b: SplunkHealth = splunk.health.publish_health()

    # Wait until we find the results in the Topic for A
    result = wait_until_topic_match(cliv1,
                                    topic="sts_health_sync",
                                    query="message.HealthSyncMessage.payload.CheckStates.checkStates[*]",
                                    contains_dict={
                                        "name": health_a.get("name"),
                                        "checkStateId": health_a.get("check_state_id"),
                                        "health": health_a.get("health"),
                                        "message": health_a.get("message"),
                                        "topologyElementIdentifier": health_a.get("topology_element_identifier"),
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=5,
                                    on_failure_action=lambda: simulator_dump())

    logging.info(f"Found the following results: {result}")

    # Wait until we find the results in the Topic for B
    result = wait_until_topic_match(cliv1,
                                    topic="sts_health_sync",
                                    query="message.HealthSyncMessage.payload.CheckStates.checkStates[*]",
                                    contains_dict={
                                        "name": health_b.get("name"),
                                        "checkStateId": health_b.get("check_state_id"),
                                        "health": health_b.get("health"),
                                        "message": health_b.get("message"),
                                        "topologyElementIdentifier": health_b.get("topology_element_identifier"),
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
def test_splunk_health_stateful_state(agent: AgentTestingBase,
                                      cliv1: CLIv1,
                                      splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_health(expect_failure: bool = False,
                    expected_health: SplunkHealth = None) -> Optional[SplunkHealth]:
        health: SplunkHealth = expected_health

        if health is None:
            health = splunk.health.publish_health()

        try:
            # Wait until we find the results in the Topic
            wait_until_topic_match(cliv1,
                                   topic="sts_health_sync",
                                   query="message.HealthSyncMessage.payload.CheckStates.checkStates[*]",
                                   contains_dict={
                                       "name": health.get("name"),
                                       "checkStateId": health.get("check_state_id"),
                                       "health": health.get("health"),
                                       "message": health.get("message"),
                                       "topologyElementIdentifier": health.get("topology_element_identifier"),
                                   },
                                   first_match=True,
                                   timeout=180,
                                   period=5)
        except Exception as e:
            if expect_failure is True:
                return health
            else:
                raise e

        if expect_failure is True:
            raise Exception("Health should not exist but did not fail with a exception")
        else:
            return health

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    health_posted_while_agent_was_down: Optional[SplunkHealth] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_health_while_agent_is_stopped():
        nonlocal health_posted_while_agent_was_down
        health_posted_while_agent_was_down = post_health(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_health_after_agent_started():
        post_health(expected_health=health_posted_while_agent_was_down)
        # Post extra health to make sure new ones also work
        post_health()

    # Run a stateful test for the agent
    agent.stateful_state_run_cycle_test(
        func_before_agent_stop=post_health,
        func_after_agent_stop=find_health_while_agent_is_stopped,
        func_after_agent_startup=find_health_after_agent_started
    )


# Transactional State
# We will produce health while the routing is open
# Then we will close the routes, post another health state and make sure that the health does not exist
# After that we will open the routes and test if the health eventually end up in STS
@pytest.mark.order(4)
def test_splunk_health_transactional_check(agent: AgentTestingBase,
                                           cliv1: CLIv1,
                                           splunk: SplunkBase,
                                           simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_health(expect_failure: bool = False, expected_health: SplunkHealth = None) -> Optional[SplunkHealth]:
        health: SplunkHealth = expected_health

        if health is None:
            health = splunk.health.publish_health()

        try:
            # Wait until we find the results in the Topic
            wait_until_topic_match(cliv1,
                                   topic="sts_health_sync",
                                   query="message.HealthSyncMessage.payload.CheckStates.checkStates[*]",
                                   contains_dict={
                                       "name": health.get("name"),
                                       "checkStateId": health.get("check_state_id"),
                                       "health": health.get("health"),
                                       "message": health.get("message"),
                                       "topologyElementIdentifier": health.get("topology_element_identifier"),
                                   },
                                   first_match=True,
                                   timeout=180,
                                   period=5)
        except Exception as e:
            if expect_failure is True:
                return health
            else:
                raise e

        if expect_failure is True:
            raise Exception("Health should not exist but did not fail with a exception")
        else:
            return health

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    health_posted_while_agent_was_down: Optional[SplunkHealth] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_health_while_routes_is_blocked():
        nonlocal health_posted_while_agent_was_down
        health_posted_while_agent_was_down = post_health(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_health_while_routes_is_open():
        post_health(expected_health=health_posted_while_agent_was_down)
        # Post extra health to make sure new ones also work
        post_health()

    # Run a stateful test for the agent
    agent.transactional_run_cycle_test(
        func_before_blocking_routes=post_health,
        func_after_blocking_routes=find_health_while_routes_is_blocked,
        rerun_func_unblocking_blocked_routes=find_health_while_routes_is_open
    )
