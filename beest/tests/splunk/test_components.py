import pytest
import util

from typing import Optional
from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from splunk_testing_base import SplunkBase, SplunkTopologyComponent
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


@pytest.mark.order(1)
def test_splunk_component(agent: AgentTestingBase,
                          splunk: SplunkBase,
                          cliv1: CLIv1,
                          simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    component: SplunkTopologyComponent = splunk.topology.publish_component()

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher() \
            .component(component.get("id"), name=component.get("id"), type=component.get("type"))

    topology_query = f"name = '{component.get('id')}'"

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: topology_query,
        timeout=80,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator_dump()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    print(f"Found the following topology query: {topology_query}")


@pytest.mark.order(2)
def test_splunk_multiple_component(agent: AgentTestingBase,
                                   splunk: SplunkBase,
                                   cliv1: CLIv1,
                                   simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    component_a: SplunkTopologyComponent = splunk.topology.publish_component()
    component_b: SplunkTopologyComponent = splunk.topology.publish_component()
    component_c: SplunkTopologyComponent = splunk.topology.publish_component()

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher() \
            .component(component_a.get("id"), name=component_a.get("id"), type=component_a.get("type")) \
            .component(component_b.get("id"), name=component_b.get("id"), type=component_b.get("type")) \
            .component(component_c.get("id"), name=component_c.get("id"), type=component_c.get("type"))

    topology_query = f"name = '{component_a.get('id')}' OR " \
                     f"name = '{component_b.get('id')}' OR " \
                     f"name = '{component_c.get('id')}'"

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: topology_query,
        timeout=80,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator_dump()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    print(f"Found the following topology query: {topology_query}")
