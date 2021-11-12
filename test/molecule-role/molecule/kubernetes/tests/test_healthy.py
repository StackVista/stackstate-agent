import os
import util
import pytest
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('kubernetes-cluster-agent')

kubeconfig_env = "KUBECONFIG=/home/ubuntu/deployment/aws-eks/tf-cluster/kubeconfig "


@pytest.mark.first
def test_node_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")

    def assert_healthy():
        c = kubeconfig_env + "kubectl wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=agent pod --namespace={}".format(namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)


@pytest.mark.second
def test_cluster_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")

    def assert_healthy():
        c = kubeconfig_env + "kubectl wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=cluster-agent pod --namespace={}".format(namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)
