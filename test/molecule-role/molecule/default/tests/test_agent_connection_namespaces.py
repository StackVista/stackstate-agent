import os
import re
import util
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('agent_linux_connection_namespace_vm')

def test_stackstate_agent_is_installed(host):
    agent = host.package("stackstate-agent")
    print agent.version
    assert agent.is_installed

    agent_current_branch = host.ansible("include_vars", "./common_vars.yml")["ansible_facts"]["agent_current_branch"]
    if agent_current_branch is "master":
        assert agent.version.startswith("2")


def test_stackstate_agent_running_and_enabled(host):
    assert not host.ansible("service", "name=stackstate-agent enabled=true state=started")['changed']


def test_stackstate_process_agent_running_and_enabled(host):
    # We don't check enabled because on systemd redhat is not needed check omnibus/package-scripts/agent/posttrans
    assert not host.ansible("service", "name=stackstate-agent-process state=started", become=True)['changed']

def test_etc_docker_directory(host):
    f = host.file('/etc/docker/')
    assert f.is_directory


def test_docker_compose_file(host):
    f = host.file('/home/ubuntu/docker-compose.yml')
    assert f.is_file
