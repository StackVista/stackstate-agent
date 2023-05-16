from packaging import version
import pytest

from conftest import STS_CONTEXT_FILE
from ststest import TopologyMatcher

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def test_projected_volume_topology(ansible_var, cliv1):
    k8s_version = ansible_var("agent_k8s_version")

    if version.parse(k8s_version) >= version.parse("1.21"):
        namespace = ansible_var("monitoring_namespace")
        release_name = ansible_var("agent_release_name")

        cluster_agent = release_name + "-cluster-agent"

        expected_topology = TopologyMatcher() \
            .component("cluster-agent", type="pod", name=fr"{cluster_agent}-\w{{7,10}}-\w{{5}}") \
            .component("cluster-agent-container", type="container", name="cluster-agent") \
            .component("kube-api-access", type="volume", name=r"kube-api-access-.*") \
            .component("kube-root-ca", type="configmap", name="kube-root-ca.crt") \
            .one_way_direction("cluster-agent", "kube-api-access", type="claims") \
            .one_way_direction("cluster-agent-container", "kube-api-access", type="mounts") \
            .one_way_direction("kube-api-access", "kube-root-ca", type="projects")

        current_topology = cliv1.topology(f"label IN ('namespace:{namespace}')", "projected-volume",
                                          config_location=STS_CONTEXT_FILE)
        possible_matches = expected_topology.find(current_topology)
        matched_res = possible_matches.assert_exact_match()

        assert 'kind:projection' in matched_res.component("kube-api-access").tags
    else:
        pytest.skip("volume projection not available before k8s 1.21")
