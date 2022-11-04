import logging
import pytest
import os
import shutil
import testinfra.utils.ansible_runner

from agent_tesing_base import AgentTestingBase
from typing import Callable
from stscliv1 import CLIv1

USE_CACHE = False

# Directory to the yard that will be used for inventory and configuration
YARD_LOCATION = f"../../sut/yards/agent-upgrade"


# Create an Agent Testing Base class to group all the functionality for the agent
def get_agent_interface(agent_os_target) -> AgentTestingBase:
    # Open up the ansible_inventory inventory again based on the same one we created the testinfra_hosts with
    ansible_inventory = testinfra.utils.ansible_runner.AnsibleRunner(f'{YARD_LOCATION}/ansible_inventory')

    # Now we select the other host, not local
    agent_ubuntu_variables = ansible_inventory.get_variables(f"agent_{agent_os_target}")

    agent_ubuntu_host = agent_ubuntu_variables.get(f"agent_{agent_os_target}")["host"]
    agent_ubuntu_user = agent_ubuntu_variables.get(f"agent_{agent_os_target}")["user"]

    agent_interface = AgentTestingBase(ansible_var,
                                       hostname=agent_ubuntu_host,
                                       username=agent_ubuntu_user,
                                       key_file_path=f'{YARD_LOCATION}/agent_{agent_os_target}_id_rsa')

    return agent_interface


@pytest.fixture
def agent_ubuntu(ansible_var, host) -> AgentTestingBase:
    agent_interface = get_agent_interface("ubuntu")
    yield agent_interface
    agent_interface.close_connection()


@pytest.fixture
def agent_redhat(ansible_var, host) -> AgentTestingBase:
    agent_interface = get_agent_interface("redhat")
    yield agent_interface
    agent_interface.close_connection()


# Copy and cleanup all the necessary files to be able to run the cli.
# This will copy the conf.yaml that was generated for StackState to the root directory the tests run in.
# Allowing the cli v1 to use this conf.d directory to successfully connect to the StackState instance
# Without this if you run the python scripts locally it will fail to retrieve data from the StackState instance
# The cli has functionality to load a conf file if there is a conf.d folder with a conf.yaml file inside
@pytest.fixture
def cliv1_configure() -> None:
    file = "conf.yaml"
    src = f"{YARD_LOCATION}"
    dest = "./conf.d"

    # Create the conf directory if it does not exist with the current test directory
    if not os.path.exists(dest):
        os.makedirs(dest)

    # Copy the conf.yaml file containing the information to connect to StackState a the created conf.d folder
    shutil.copy(f"{src}/{file}", f"{dest}/{file}")

    # We can yield on this point allowing the code to continue and use the conf.yaml file, we will clean this afterwards
    yield

    # Remove the conf.d folder after the cycle has run to make sure we do not leave and zombie file that may interfere
    # and give a false positive, Doubt this will happen but to be sure
    shutil.rmtree(dest)


# Create a cli v1 instance that can be used to query StackState
@pytest.fixture
def cliv1(cliv1_configure, host, caplog) -> CLIv1:
    caplog.set_level(logging.INFO)
    return CLIv1(host, log=logging.getLogger("CLIv1"),
                 cache_enabled=USE_CACHE)


# Retrieve an ansible variable from the Ansible Inventory based on the host
# This is based on the host your selected in the testinfra_hosts variable
@pytest.fixture
def ansible_var(host) -> Callable[[str], str]:
    def _retrieve_var(name, none_on_undefined=False):
        raw_vars = host.ansible.get_variables()
        if name in raw_vars and (type(raw_vars[name]) != str or "{{" not in raw_vars[name]):
            # No interpolation needed, we return the raw value
            return raw_vars[name]
        else:
            if none_on_undefined:
                return None

            # This allows variable interpolation
            # https://stackoverflow.com/questions/57820998/accessing-ansible-variables-in-molecule-test-testinfra
            return host.ansible("debug", "msg={{ " + name + " }}")["msg"]
    return _retrieve_var

