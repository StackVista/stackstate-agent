import random
import util

from produce_data import publish_components
from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_splunk_server_component(ansible_var, splunk_instance, cliv1, host):
    yard_id = ansible_var("yard_id")

    # Generated Values For Random Data
    component_id = "server_{}".format(random.randint(0, 10000))

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    publish_components(ansible_var, splunk_instance, component_id)

    # The component may not appear instantly thus we have a wait with a timeout
    def wait_for_component():
        expected_topology = TopologyMatcher() \
            .component(component_id, type="server")

        topology_query = "label IN ('stackpack:splunk', 'splunk-instance:{}') AND name = '{}'"\
            .format(yard_id, component_id)

        current_topology = cliv1.topology(topology_query)
        possible_matches = expected_topology.find(current_topology)
        possible_matches.assert_exact_match()

    util.wait_until(wait_for_component, 60, 5)
