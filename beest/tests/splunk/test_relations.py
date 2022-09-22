import util
import random

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
    # Component A
    component_id_source = "server_{}".format(random.randint(0, 10000))
    component_type = "server"
    component_description = "Topology Server Component"

    splunk.topology.publish_component(component_id=component_id_source,
                                      component_type=component_type,
                                      description=component_description)

    # Component B
    component_id_target = "server_{}".format(random.randint(0, 10000))
    component_type = "server"
    component_description = "Topology Server Component"

    splunk.topology.publish_component(component_id=component_id_target,
                                      component_type=component_type,
                                      description=component_description)

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology.publish_relation(relation_type="CONNECTED",
                                     source_id=component_id_source,
                                     target_id=component_id_target)

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher()\
            .component(component_id_source, name=component_id_source, type="server")\
            .component(component_id_target, name=component_id_target, type="server")\
            .one_way_direction(source=component_id_source, target=component_id_target)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{component_id_source}' OR name = '{component_id_target}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )
