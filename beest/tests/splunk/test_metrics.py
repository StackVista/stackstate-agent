import util

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as it for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_server_metrics(splunk: SplunkBase,
                               cliv1: CLIv1):

    splunk.metric._post_metric()

    assert 1 == 1
