import random
import logging

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_events(splunk: SplunkBase,
                       cliv1: CLIv1,
                       simulator):
    # Prepare the data that will be sent to StackState
    random_host_id = random.randint(0, 10000)

    event_host = "host{}".format(random_host_id)
    event_source_type = "sts_test_data"
    event_status = random.choice(["CRITICAL", "OK", "ERROR", "WARNING"])
    event_description = "Test host{} Event".format(random_host_id)

    # Post the health data to StackState
    splunk.event.publish_event(host=event_host,
                               source_type=event_source_type,
                               status=event_status,
                               description=event_description)

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_generic_events",
                                    query="message.GenericEvent.tags",
                                    contains_dict={
                                        "source_type_name": "generic_splunk_event",
                                        "host": event_host,
                                        "description": event_description,
                                        "status": event_status
                                    },
                                    first_match=True,
                                    timeout=180,
                                    period=10,
                                    on_failure_action=lambda: simulator())

    logging.info(f"Found the following results: {result}")
