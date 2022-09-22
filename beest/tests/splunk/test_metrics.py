import logging
import util
import random

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_metrics(splunk: SplunkBase,
                        cliv1: CLIv1,
                        simulator):
    # Prepare the data that will be sent to StackState
    random_host_id = random.randint(0, 10000)

    metric_host = "host{}".format(random_host_id)
    metric_source_type = "sts_test_data"
    metric_topo_type = "metrics"
    metric_id = "raw.metrics"
    metric_value = random.randint(1, 100000)
    metric_qa = "splunk"

    splunk.metric.publish_metric(host=metric_host,
                                 source_type=metric_source_type,
                                 topo_type=metric_topo_type,
                                 metric=metric_id,
                                 value=metric_value,
                                 qa=metric_qa)

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_multi_metrics",
                                    query="message.MultiMetric.values",
                                    contains_dict={
                                        "raw.metrics": metric_value,
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=10,
                                    on_failure_action=lambda: simulator())

    logging.info(f"Found the following results: {result}")
