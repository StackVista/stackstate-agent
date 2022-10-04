import logging
from typing import Optional

import pytest

from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from splunk_testing_base import SplunkBase, SplunkHealth
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
                                    timeout=120,
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
                                    timeout=120,
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
                                    timeout=120,
                                    period=5,
                                    on_failure_action=lambda: simulator_dump())

    logging.info(f"Found the following results: {result}")
