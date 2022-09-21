import util
import time

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_server_metrics(splunk: SplunkBase,
                               cliv1: CLIv1,
                               simulator):

    splunk.metric._post_metric()

    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics")

        metrics = {}
        for message in json_data["messages"]:
            for m_name in message["message"]["MultiMetric"]["values"].keys():
                if m_name not in metrics:
                    metrics[m_name] = []

                values = [message["message"]["MultiMetric"]["values"][m_name]]
                metrics[m_name] += values

        assert metrics["raw.metrics"][0] == 3.0

    util.wait_until(wait_for_metrics, 60, 3)
