import logging
import pytest
import os
import requests
import shutil
import json

from pathlib import Path
from splunk_testing_base import SplunkBase
from agent_tesing_base import AgentTestingBase
from typing import Callable
from stscliv1 import CLIv1

USE_CACHE = False

# Directory to the yard that will be used for inventory and configuration
YARD_LOCATION = f"../../sut/yards/splunk"


@pytest.fixture
def simulator_dump(request, ansible_var, splunk):
    def dump_data(max_results=100):
        logging.error("Dumping the StackState Simulator logs. This may take some time depending on how long the"
                      " simulator has been running for ...")

        # Retrieve if the sts simulator is enabled in the playbook scenario
        ansible_sts_simulator_var = "enable_sts_simulator"
        is_simulator_enabled = ansible_var(ansible_sts_simulator_var, True)

        # If is_simulator_enabled is enabled then we will continue with pulling the simulator data
        if is_simulator_enabled is True:
            data_dump_filename = "{}-{}-simulator_dump.json".format(
                Path(str(request.node.fspath)).stem,
                request.node.originalname,
            )

            # Attempt to pull data from the splunk host as the simulator will run on the same host
            response = requests.get(url=f"http://{splunk.splunk_host}:7078/download")
            response_data = response.json()

            # Dump the data before starting to play around with it like trimming it
            with open(data_dump_filename, "w") as outfile:
                json.dump(response_data, outfile, indent=4)

            # Trim data to be smaller when used in a debugger
            if len(response_data) > max_results:
                logging.warning(f"Received a total of {len(response_data)} results from the Simulator")
                logging.warning(f"This is more than max of {max_results} results, Trimming the array to {max_results} "
                                f"results")
                logging.warning(f"You can find the complete results in the data dump file: {data_dump_filename}")
                response_data = response_data[:max_results]

            return response_data

        elif is_simulator_enabled is None:
            logging.warning(f"Skipping: '{ansible_sts_simulator_var}' is not defined the the yard all.yml, "
                            f"Skipping a simulator data dump")
            return None

        else:
            logging.warning("Skipping: Simulator is disabled, skipping a simulator data dump")
            return None

    return dump_data


# Create a Agent Testing Base class to group all the functionality for the agent
@pytest.fixture
def agent(ansible_var, host, splunk) -> AgentTestingBase:
    agent_interface = AgentTestingBase(ansible_var,
                                       hostname=f'{splunk.splunk_host}',
                                       username=f'ubuntu',
                                       key_file_path=f'{YARD_LOCATION}/splunk_id_rsa')

    yield agent_interface

    agent_interface.close_connection()


# Create a Splunk Testing Base class to group all the functionality for splunk
# Allowing it to be used on other testing script if need be
@pytest.fixture
def splunk(ansible_var, request, host) -> SplunkBase:
    return SplunkBase(host,
                      request=request,
                      ansible_var=ansible_var,
                      yard_location=YARD_LOCATION)


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

