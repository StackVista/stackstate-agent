import logging
from platform import release
import secrets

from conftest import STS_CONTEXT_FILE
from ststest import TopologyMatcher

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def test_cluster_agent_topology(ansible_var, cliv1):
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")
    release_name = ansible_var("agent_release_name")
    branch_name = ansible_var("agent_current_branch")
    commit_sha = ansible_var("agent_current_branch_ref")

    cluster_agent = release_name + "-cluster-agent"

    if release_name == "stackstate-k8s-agent":
        secret_name = release_name
    else:
        secret_name = release_name + "-stackstate-k8s-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node", type="node", name=r"ip-.*") \
        .component("cluster-agent-deployment", type="deployment", name=cluster_agent) \
        .component("cluster-agent-rs", type="replicaset", name=fr"{cluster_agent}-\w{{7,10}}") \
        .component("cluster-agent", type="pod", name=fr"{cluster_agent}-\w{{7,10}}-\w{{5}}") \
        .component("cluster-agent-container", type="container", name="cluster-agent") \
        .component("cluster-agent-cm", type="configmap", name=cluster_agent) \
        .component("cluster-agent-secret", type="secret", name=secret_name) \
        .component("cluster-agent-svc", type="service", name=cluster_agent) \
        .component("cluster-agent-cluster-agent", type="stackstate-agent", name="stackstate-cluster-agent start") \
        .one_way_direction("cluster-agent-deployment", "cluster-agent-rs", type="controls") \
        .one_way_direction("cluster-agent-svc", "cluster-agent", type="exposes") \
        .one_way_direction("cluster-agent-rs", "cluster-agent", type="controls") \
        .one_way_direction("cluster-agent", "cluster-agent-container", type="encloses") \
        .one_way_direction("cluster-agent", "cluster-agent-cm", type="claims") \
        .one_way_direction("cluster-agent", "node", type="scheduled_on") \
        .one_way_direction("cluster-agent-container", "cluster-agent-cm", type="mounts") \
        .one_way_direction("cluster-agent", "cluster-agent-secret", type="uses_value") \
        .one_way_direction("cluster-agent-container", "cluster-agent-cluster-agent", type="runs")

    matched_res = query_and_assert(cliv1, cluster_name, namespace, expected_topology)
    # TODO revive with STAC-19236
    # assert f"image_tag:{branch_name}" in matched_res.component("cluster-agent-container").tags
    # also commit_sha


def test_node_agent_topology(ansible_var, cliv1):
    k8s_node_count = int(ansible_var("agent_k8s_size"))
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")
    release_name = ansible_var("agent_release_name")
    branch_name = ansible_var("agent_current_branch")
    commit_sha = ansible_var("agent_current_branch_ref")

    node_agent = release_name + "-node-agent"

    if release_name == "stackstate-k8s-agent":
        secret_name = release_name
    else:
        secret_name = release_name + "-stackstate-k8s-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node-agent-svc", type="service", name=node_agent) \
        .component("node-agent-ds", type="daemonset", name=node_agent) \
        .component("node-agent-cm", type="configmap", name=node_agent) \
        .component("node-agent-secret", type="secret", name=secret_name) \
        .repeated(
            k8s_node_count,
            lambda matcher: matcher
            .component("node", type="node", name=r"ip-.*")
            .component("node-agent", type="pod", name=fr"{node_agent}-.*")
            .component("node-agent-main-container", type="container", name="node-agent")
            .component("node-agent-process-container", type="container", name="process-agent")
            .component("cgroups-vol", type="volume", name="cgroups")
            .component("node-agent-process-agent", type="stackstate-agent", name="process-agent")
            .component("node-agent-trace-agent", type="stackstate-agent", name="trace-agent")
            .component("node-agent-main-agent", type="stackstate-agent", name="agent run")
            .one_way_direction("node-agent", "node", type="scheduled_on")
            .one_way_direction("node-agent-ds", "node-agent", type="controls")
            .one_way_direction("node-agent-svc", "node-agent", type="exposes")
            .one_way_direction("node-agent", "node-agent-main-container", type="encloses")
            .one_way_direction("node-agent", "node-agent-process-container", type="encloses")
            .one_way_direction("node-agent", "node-agent-cm", type="claims")
            .one_way_direction("node-agent-main-container", "node-agent-cm", type="mounts")
            .one_way_direction("node-agent-process-container", "node-agent-cm", type="mounts")
            .one_way_direction("node-agent-main-container", "cgroups-vol", type="mounts")
            .one_way_direction("node-agent-process-container", "cgroups-vol", type="mounts")
            .one_way_direction("node-agent", "node-agent-secret", type="uses_value")
            .one_way_direction("node-agent-process-container", "node-agent-process-agent", type="runs")
            .one_way_direction("node-agent-main-container", "node-agent-trace-agent", type="runs")
            .one_way_direction("node-agent-main-container", "node-agent-main-agent", type="runs")
        )

    matched_res = query_and_assert(cliv1, cluster_name, namespace, expected_topology)

    # TODO revive after STAC-19236
    # node_agent_pod_name = matched_res.component(("node-agent", 0)).name
    # assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-main-container", 0)).tags
    # assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-main-agent", 0)).tags
    # assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-trace-agent", 0)).tags
    # assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-process-container", 0)).tags
    # assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-process-agent", 0)).tags
    # assert f"image_tag:{branch_name}" in matched_res.component(("node-agent-main-container", 0)).tags
    # also commit_sha


def test_checks_agent_topology(ansible_var, cliv1):
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")
    release_name = ansible_var("agent_release_name")
    branch_name = ansible_var("agent_current_branch")
    commit_sha = ansible_var("agent_current_branch_ref")

    checks_agent = release_name + "-checks-agent"

    if release_name == "stackstate-k8s-agent":
        secret_name = release_name
    else:
        secret_name = release_name + "-stackstate-k8s-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node", type="node", name=r"ip-.*") \
        .component("checks-agent-deployment", type="deployment", name=checks_agent) \
        .component("checks-agent-rs", type="replicaset", name=fr"{checks_agent}-.*") \
        .component("checks-agent", type="pod", name=fr"{checks_agent}-.*-.*") \
        .component("checks-agent-secret", type="secret", name=secret_name) \
        .component("checks-agent-main-agent", type="stackstate-agent", name="agent run") \
        .component("checks-agent-container", type="container", name="stackstate-k8s-agent") \
        .one_way_direction("checks-agent-deployment", "checks-agent-rs", type="controls") \
        .one_way_direction("checks-agent-rs", "checks-agent", type="controls") \
        .one_way_direction("checks-agent", "node", type="scheduled_on") \
        .one_way_direction("checks-agent", "checks-agent-container", type="encloses") \
        .one_way_direction("checks-agent", "checks-agent-secret", type="uses_value") \
        .one_way_direction("checks-agent-container", "checks-agent-main-agent", type="runs")

    matched_res = query_and_assert(cliv1, cluster_name, namespace, expected_topology)
    # TODO revive after STAC-19236
    # assert f"image_tag:{branch_name}" in matched_res.component("checks-agent-container").tags
    # also commit_sha


def query_and_assert(cliv1, cluster_name: str, namespace: str, expected_topology: TopologyMatcher):
    current_agent_topology = cliv1.topology(
        f"(label IN ('cluster-name:{cluster_name}') AND label IN ('namespace:{namespace}'))"
        f" OR (type IN ('node', 'namespace', 'stackstate-agent', 'stackstate-k8s-agent'))", "agent", config_location=STS_CONTEXT_FILE)

    possible_matches = expected_topology.find(current_agent_topology)
    return possible_matches.assert_exact_match()
