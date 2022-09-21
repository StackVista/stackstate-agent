import logging
import random

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from util import wait_until_topic_match

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_server_health(splunk: SplunkBase,
                              cliv1: CLIv1,
                              simulator):
    # Prepare the data that will be sent to StackState
    random_disk_id = random.randint(0, 10000)
    random_server_id = random.randint(0, 10000)

    health_name = "Disk {} SDA".format(random_disk_id)
    health_state_id = "disk_{}_sda".format(random_disk_id)
    health_message = "SDA Disk {} Message".format(random_disk_id)
    health_health = random.choice(["CLEAR", "CRITICAL"])
    health_topo_identifier = "server_{}".format(random_server_id)

    # Post the health data to StackStat
    splunk.health._post_health(name=health_name,
                               check_state_id=health_state_id,
                               health=health_health,
                               topology_element_identifier=health_topo_identifier,
                               message=health_message)

    # Wait until we find the results in the Topic
    result = wait_until_topic_match(cliv1,
                                    topic="sts_health_sync",
                                    query="message.HealthSyncMessage.payload.CheckStates.checkStates[0]",
                                    contains_dict={
                                        "name": health_name,
                                        "checkStateId": health_state_id,
                                        "health": health_health,
                                        "message": health_message,
                                        "topologyElementIdentifier": health_topo_identifier,
                                    },
                                    first_match=True,
                                    timeout=120,
                                    period=10,
                                    on_failure_action=lambda: simulator())

    logging.info(f"Found the following results: {result}")
