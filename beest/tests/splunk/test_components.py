import util

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as it for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_server_component(splunk: SplunkBase,
                                 cliv1: CLIv1):

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    component_id = splunk.topology.publish_random_server_component()

    # Wait until we find this component in StackState
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=lambda:
            TopologyMatcher()
            .component("random-server-component", name=component_id, type="server"),
        topology_query=lambda:
            f"name = '{component_id}'",
        timeout=60,
        period=5
    )
