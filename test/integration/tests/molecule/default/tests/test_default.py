import os

import testinfra.utils.ansible_runner

testinfra_hosts = testinfra.utils.ansible_runner.AnsibleRunner(
    os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('all')


# EC2 provides unique random hostnames.
def test_hostname(host):
    pass


def test_opt_stackstate_directory(host):
    f = host.file('/opt/datadog-agent/')

    assert f.is_directory


def test_etc_docker_directory(host):
    f = host.file('/etc/docker/')

    assert f.is_directory


def test_docker_compose_file(host):
    f = host.file('/etc/docker-compose.yml')

    assert f.is_file
