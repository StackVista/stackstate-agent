import util
import pytest

testinfra_hosts = ["local"]


@pytest.mark.order(1)
def test_node_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")
    kubecontext = ansible_var("agent_kubecontext")

    def assert_healthy():
        c = "kubectl --context={0} wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=agent pod --namespace={1}".format(kubecontext, namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)


@pytest.mark.order(2)
def test_cluster_agent_healthy(host, ansible_var):
    namespace = ansible_var("namespace")
    kubecontext = ansible_var("agent_kubecontext")

    def assert_healthy():
        c = "kubectl --context={0} wait --for=condition=ready --timeout=1s -l app.kubernetes.io/component=cluster-agent pod --namespace={1}".format(kubecontext, namespace)
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 30, 5)
