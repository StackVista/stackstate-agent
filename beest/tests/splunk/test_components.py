import logging
import util
import random
import time
import paramiko

from splunk_testing_base import SplunkBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_component(splunk: SplunkBase,
                          cliv1: CLIv1,
                          simulator):
    # Prepare the data that will be sent to StackState
    component_id = "server_{}".format(random.randint(0, 10000))
    component_type = "server"
    component_description = "Topology Server Component"

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    splunk.topology.publish_component(component_id=component_id,
                                      component_type=component_type,
                                      description=component_description)

    logging.debug(f"Attempting to find a component with the name '{component_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher()\
            .component(component_id, name=component_id, type=component_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{component_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


def test_splunk_multiple_component(splunk: SplunkBase,
                                   cliv1: CLIv1,
                                   simulator):
    # Component A
    component_a_id = "server_{}".format(random.randint(0, 10000))
    component_a_type = "server"
    component_a_description = f"Topology Server Component A {component_a_id}"

    splunk.topology.publish_component(component_id=component_a_id, component_type=component_a_type,
                                      description=component_a_description)

    # Component B
    component_b_id = "server_{}".format(random.randint(0, 10000))
    component_b_type = "server"
    component_b_description = f"Topology Server Component B {component_b_id}"

    splunk.topology.publish_component(component_id=component_b_id, component_type=component_b_type,
                                      description=component_b_description)

    # Component C
    component_c_id = "server_{}".format(random.randint(0, 10000))
    component_c_type = "server"
    component_c_description = f"Topology Server Component C {component_c_id}"

    splunk.topology.publish_component(component_id=component_c_id, component_type=component_c_type,
                                      description=component_c_description)

    logging.debug(f"Attempting to find a component with the name '{component_a_id}' on StackState")
    logging.debug(f"Attempting to find a component with the name '{component_b_id}' on StackState")
    logging.debug(f"Attempting to find a component with the name '{component_c_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher() \
            .component(component_a_id, name=component_a_id, type=component_a_type) \
            .component(component_b_id, name=component_b_id, type=component_b_type) \
            .component(component_c_id, name=component_c_id, type=component_c_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{component_a_id}' OR name = '{component_b_id}' OR name = '{component_c_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


# Stateful State
# We will publish a component while the agent is active and wait for it
# When we find it then we will stop the agent and post a second component
# After a few minutes we start the agent up again
# And wait to find the second component
def test_component_stateful_state(cliv1: CLIv1, simulator, splunk: SplunkBase):
    # SSH Connection to the Splunk host
    # This can allow us to later on start and stop the StackState Agent
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=f'{splunk.splunk_host}',
                   username='ubuntu',
                   key_filename=f'{YARD_LOCATION}/splunk_id_rsa')

    def start_agent():
        logging.info("Starting the stackstate-agent agent ...")
        client.exec_command('sudo service stackstate-agent start')

        # Give the agent 10 seconds to start up
        time.sleep(10)

        # Test if the agent is running
        stdin, stdout, stderr = client.exec_command('sudo systemctl is-active --quiet stackstate-agent && '
                                                    'echo Service is running')
        if stdout.read() == b'':
            raise ProcessLookupError("The Agent is not running on the host machine. The Agent first needs to run")

    # First we make sure the agent is running, if the test for some reason failed half way the agent may be turned off
    # for testing purposes
    start_agent()

    # Publish the first component before stopping the agent
    before_stop_component_id = "server_{}".format(random.randint(0, 10000))
    before_stop_component_type = "server"
    before_stop_component_description = "Topology Server Component"

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    splunk.topology.publish_component(component_id=before_stop_component_id,
                                      component_type=before_stop_component_type,
                                      description=before_stop_component_description)

    logging.info(f"Attempting to find a component with the name '{before_stop_component_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher().component(before_stop_component_id,
                                           name=before_stop_component_id,
                                           type=before_stop_component_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{before_stop_component_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    # Stop the agent so that we can test stateful state
    logging.info("Stopping the stackstate-agent agent ...")
    client.exec_command('sudo service stackstate-agent stop')

    # Give the agent 5 seconds to stop
    time.sleep(5)

    # Test if the agent is stopped
    stdin, stdout, stderr = client.exec_command('sudo systemctl is-active --quiet stackstate-agent && '
                                                'echo Service is running')
    if stdout.read() != b'':
        raise ProcessLookupError("Unable to stop the Agent, The agent needs to be stopped to test stateful")

    # Publish a second component before starting the agent
    durning_stop_component_id = "server_{}".format(random.randint(0, 10000))
    durning_stop_component_type = "server"
    durning_stop_component_description = "Topology Server Component"

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    splunk.topology.publish_component(component_id=durning_stop_component_id,
                                      component_type=durning_stop_component_type,
                                      description=durning_stop_component_description)

    # Sleeping for two minutes to give more than enough time to pass before testing state
    logging.info("Waiting 2 min to have a larger gap before starting the Agent up...")

    time.sleep(120)
    start_agent()

    time.sleep(10)
    client.close()

    logging.info(f"Attempting to find a component with the name '{durning_stop_component_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def durning_stop_topology_matcher():
        return TopologyMatcher().component(durning_stop_component_id,
                                           name=durning_stop_component_id,
                                           type=durning_stop_component_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=durning_stop_topology_matcher,
        topology_query=lambda: f"name = '{durning_stop_component_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


def test_component_should_not_exist_before_agent(cliv1: CLIv1, simulator, splunk: SplunkBase):
    # SSH Connection to the Splunk host
    # This can allow us to later on start and stop the StackState Agent
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=f'{splunk.splunk_host}',
                   username='ubuntu',
                   key_filename=f'{YARD_LOCATION}/splunk_id_rsa')

    def start_agent():
        logging.info("Starting the stackstate-agent agent ...")
        client.exec_command('sudo service stackstate-agent start')

        # Give the agent 10 seconds to start up
        time.sleep(10)

        # Test if the agent is running
        stdin, stdout, stderr = client.exec_command('sudo systemctl is-active --quiet stackstate-agent && '
                                                    'echo Service is running')
        if stdout.read() == b'':
            raise ProcessLookupError("The Agent is not running on the host machine. The Agent first needs to run")

    # First we make sure the agent is running, if the test for some reason failed half way the agent may be turned off
    # for testing purposes
    start_agent()

    # Publish the first component before stopping the agent
    before_stop_component_id = "server_{}".format(random.randint(0, 10000))
    before_stop_component_type = "server"
    before_stop_component_description = "Topology Server Component"

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    splunk.topology.publish_component(component_id=before_stop_component_id,
                                      component_type=before_stop_component_type,
                                      description=before_stop_component_description)

    logging.info(f"Attempting to find a component with the name '{before_stop_component_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher().component(before_stop_component_id,
                                           name=before_stop_component_id,
                                           type=before_stop_component_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{before_stop_component_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    # Stop the agent so that we can test stateful state
    logging.info("Stopping the stackstate-agent agent ...")
    client.exec_command('sudo service stackstate-agent stop')

    # Give the agent 5 seconds to stop
    time.sleep(5)

    # Test if the agent is stopped
    stdin, stdout, stderr = client.exec_command('sudo systemctl is-active --quiet stackstate-agent && '
                                                'echo Service is running')
    if stdout.read() != b'':
        raise ProcessLookupError("Unable to stop the Agent, The agent needs to be stopped to test stateful")

    # Publish a second component before starting the agent
    durning_stop_component_id = "server_{}".format(random.randint(0, 10000))
    durning_stop_component_type = "server"
    durning_stop_component_description = "Topology Server Component"

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    splunk.topology.publish_component(component_id=durning_stop_component_id,
                                      component_type=durning_stop_component_type,
                                      description=durning_stop_component_description)

    # Sleeping for two minutes to give more than enough time to pass before testing state
    logging.info("Waiting 2 min to have a larger gap before starting the Agent up...")
    time.sleep(120)

    # Start the agent to test Stateful
    start_agent()

    # Give the agent 10 seconds to start up
    time.sleep(10)

    logging.info(f"Attempting to find a component with the name '{durning_stop_component_id}' on StackState")

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def durning_stop_topology_matcher():
        return TopologyMatcher().component(durning_stop_component_id,
                                           name=durning_stop_component_id,
                                           type=durning_stop_component_type)

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=durning_stop_topology_matcher,
        topology_query=lambda: f"name = '{durning_stop_component_id}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )

    client.close()
