import util
import pytest

testinfra_hosts = ["local"]


@pytest.mark.order(1)
def test_node_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")
    kubeconfig = ansible_var("kubeconfig")
    kubecontext = ansible_var("kubecontext")

    def assert_healthy():
        c = "KUBECONFIG={0} kubectl --context={1} wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=agent pod --namespace={2}".format(kubeconfig, kubecontext, namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)


@pytest.mark.order(2)
def test_cluster_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")
    kubeconfig = ansible_var("kubeconfig")
    kubecontext = ansible_var("kubecontext")

    def assert_healthy():
        c = "KUBECONFIG={0} kubectl --context={1} wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=cluster-agent pod --namespace={2}".format(kubeconfig, kubecontext, namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)
