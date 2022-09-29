import util

from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from splunk_testing_base import SplunkBase, SplunkTopologyComponent
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_agent_v1_component(agent: AgentTestingBase,
                                   splunk: SplunkBase,
                                   cliv1: CLIv1,
                                   simulator_dump):
    # Allow routing to make sure we can send things to stackstate
    agent.allow_routing_to_sts_instance()

    # Stop the v2 agent
    agent.stop_agent_on_host()

    # Remove agent v2 cache
    agent.remove_agent_run_cache()

    # convert_agent_v1_run_cache_to_v2

    # Publish a v1 component, metric and event
    component: SplunkTopologyComponent = splunk.topology.publish_component(version="v1")

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
        timeout=320,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator_dump()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    print(f"Found the following topology query: {topology_query}")
