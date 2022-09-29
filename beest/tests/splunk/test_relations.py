import util
import pytest

from typing import Optional
from agent_tesing_base import AgentTestingBase
from splunk_testing_base import SplunkBase, SplunkTopologyComponent, SplunkTopologyRelation
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


@pytest.mark.order(1)
def test_splunk_relations(splunk: SplunkBase,
                          cliv1: CLIv1,
                          agent: AgentTestingBase,
                          simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    source: SplunkTopologyComponent = splunk.topology.publish_component()
    target: SplunkTopologyComponent = splunk.topology.publish_component()

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target.get("id"))

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def relation_matcher():
        return TopologyMatcher()\
            .component(source.get("id"), name=source.get("id"), type="server")\
            .component(target.get("id"), name=target.get("id"), type="server")\
            .one_way_direction(source=source.get("id"), target=target.get("id"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=relation_matcher,
        topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator_dump()  # Dump the simulator logs if the cycle failed (If enabled)
    )


@pytest.mark.order(2)
def test_splunk_multiple_relation(splunk: SplunkBase,
                                  agent: AgentTestingBase,
                                  cliv1: CLIv1,
                                  simulator_dump):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    source: SplunkTopologyComponent = splunk.topology.publish_component()
    target_a: SplunkTopologyComponent = splunk.topology.publish_component()
    target_b: SplunkTopologyComponent = splunk.topology.publish_component()

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target_a.get("id"))
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target_b.get("id"))

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def relation_matcher():
        return TopologyMatcher()\
            .component(source.get("id"), name=source.get("id"), type="server")\
            .component(target_a.get("id"), name=target_a.get("id"), type="server")\
            .component(target_b.get("id"), name=target_b.get("id"), type="server")\
            .one_way_direction(source=source.get("id"), target=target_a.get("id"))\
            .one_way_direction(source=source.get("id"), target=target_b.get("id"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=relation_matcher,
        topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target_a.get('id')}' "
                               f"OR name = '{target_b.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator_dump()  # Dump the simulator logs if the cycle failed (If enabled)
    )
