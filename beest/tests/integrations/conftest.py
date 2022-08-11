import logging
import pytest
from stscliv1 import CLIv1

USE_CACHE = False


@pytest.fixture
def cliv1(host, caplog) -> CLIv1:
    caplog.set_level(logging.INFO)
    return CLIv1(host, log=logging.getLogger("CLIv1"), cache_enabled=USE_CACHE)

@pytest.fixture
def hostname(host):
    return host.ansible.get_variables()["inventory_hostname"]

@pytest.fixture
def agent_hostname(host):
    agentHost = host.ansible.get_variables().get('groups', {}).get('agent', [])

    if agentHost:
        return agentHost[0]
    else:
        return "agent-host-name-not-found-ansible-vars"