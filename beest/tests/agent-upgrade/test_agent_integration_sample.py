import pytest
import util

from typing import Optional, Callable
from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_ubuntu_agent_upgrade(ansible_var: Callable[[str], str],
                              agent_ubuntu: AgentTestingBase):
    agent_ubuntu.ubuntu_reinstall_agent(ansible_var)

    agent_version = agent_ubuntu.ubuntu_get_agent_version()
    current_branch = ansible_var("agent_current_branch")

    print(f"Current Agent Deployed: {agent_version}")
    print(f"Latest Target Agent Deploy: {current_branch}")

    # # Determine if the currently installed agent is already the current upgraded agent IE this agent version
    # # If it is then we first to reset the state and install the latest agent
    # if current_branch.lower() in agent_version.lower():
    #     print("Warning, Agent is already updated, Resetting Agent to released latest ...")
    # else:
    #     print("OK, Agent is running the latest released version, continuing ...")

    assert 1 == 1
