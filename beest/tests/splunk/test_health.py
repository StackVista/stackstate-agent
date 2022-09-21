
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
    result = wait_until_topic_match(cliv1,
                                    topic="sts_health_sync",
                                    query="message.HealthSyncMessage",
                                    first_match=True,
                                    contains_dict={
                                        "messageId": "ddb736e1-d7de-42c5-9a56-593e6e806981"
                                    },
                                    timeout=60,
                                    period=5,
                                    on_failure_action=lambda: simulator())

    print(f"result: {result}")

    # splunk.health._post_health()
    # time.sleep(60)
    # def wait_for_metrics():
    #     json_data = cliv1.topic_api("sts_health_sync")
    #
    #     find(
    #         cliv1,
    #         topic_name="sts_health_sync"
    #     )
    #
    #     # ["message"]["HealthSyncMessage"]["payload"]["CheckState"]["checkStates"]
    #
    #     # checkStateId = 'disk_sda'
    #     # health = 'CLEAR'
    #     # message = 'sda message'
    #     # name = 'Disk sda'
    #     # topologyElementIdentifier = 'server_1'
    #
    #
    #     metrics = {}
    #     for message in json_data["messages"]:
    #         for m_name in message["message"]["HealthSyncMessage"]["payload"]["CheckState"]["checkStates"].keys():
    #             if m_name not in metrics:
    #                 metrics[m_name] = []
    #
    #             values = [message["message"]["HealthSyncMessage"]["payload"]["CheckState"]["checkStates"][m_name]]
    #             metrics[m_name] += values
    #
    #     assert metrics["raw.metrics"][0] == 3.0
    #
    # util.wait_until(wait_for_metrics, 60, 3)
