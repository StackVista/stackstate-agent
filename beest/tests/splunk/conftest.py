import logging
import pytest
import os
import shutil

from splunk_testing_base import SplunkTestingBase
from typing import Callable
from stscliv1 import CLIv1

USE_CACHE = False

# Directory to the yard that will be used for inventory and configuration
YARD_LOCATION = f"../../sut/yards/splunk"


# Create a Splunk Testing Base class to group all the functionality for splunk
# Allowing it to be used on other testing script if need be
@pytest.fixture
def splunk(ansible_var, host) -> SplunkTestingBase:
    return SplunkTestingBase(host,
                             ansible_var=ansible_var,
                             log=logging.getLogger("Splunk"),
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
    def _retrieve_var(name):
        raw_vars = host.ansible.get_variables()
        if name in raw_vars and (type(raw_vars[name]) != str or "{{" not in raw_vars[name]):
            # No interpolation needed, we return the raw value
            return raw_vars[name]
        else:
            # This allows variable interpolation
            # https://stackoverflow.com/questions/57820998/accessing-ansible-variables-in-molecule-test-testinfra
            return host.ansible("debug", "msg={{ " + name + " }}")["msg"]
    return _retrieve_var

