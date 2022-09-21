import util
import logging

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_server_relation(splunk: SplunkBase,
                                cliv1: CLIv1,
                                simulator):
    # Publish a Splunk Component to the Splunk Instance to be used in testing
    source_component_id = splunk.topology.publish_random_server_component()
    target_component_id = splunk.topology.publish_random_server_component()

    logging.debug(f"Attempting to find a source component with the name '{source_component_id}' on StackState")
    logging.debug(f"Attempting to find a target component with the name '{target_component_id}' on StackState")

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology._post_relation(relation_type="CONNECTED",
                                   source_id=source_component_id,
                                   target_id=target_component_id)

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher()\
            .component("random-server-component-source", name=source_component_id, type="server")\
            .component("random-server-component-target", name=target_component_id, type="server")\
            .one_way_direction(source="random-server-component-source",
                               target="random-server-component-target")

    # The topology_query process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_query():
        return f"name = '{source_component_id}' OR name = '{target_component_id}'"

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=topology_query,
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )
