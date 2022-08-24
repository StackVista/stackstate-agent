from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_cluster_agent_topology(ansible_var, cliv1):
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")

    cluster_agent = "stackstate-cluster-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node", type="node", name=r"ip-.*") \
        .component("cluster-agent-svc", type="service", name=cluster_agent) \
        .component("cluster-agent-deployment", type="deployment", name=cluster_agent) \
        .component("cluster-agent-rs", type="replicaset", name=fr"{cluster_agent}-\w{{9,10}}") \
        .component("cluster-agent", type="pod", name=fr"{cluster_agent}-\w{{9,10}}-\w{{5}}") \
        .component("cluster-agent-container", type="container", name="cluster-agent") \
        .component("cluster-agent-cm", type="configmap", name=cluster_agent) \
        .component("cluster-agent-secret", type="secret", name=cluster_agent) \
        .component("cluster-agent-cluster-agent", type="stackstate-agent", name="stackstate-cluster-agent start") \
        .one_way_direction("namespace", "cluster-agent-svc", type="encloses") \
        .one_way_direction("namespace", "cluster-agent-deployment", type="encloses") \
        .one_way_direction("cluster-agent-deployment", "cluster-agent-rs", type="controls") \
        .one_way_direction("cluster-agent-rs", "cluster-agent", type="controls") \
        .one_way_direction("cluster-agent-svc", "cluster-agent", type="exposes") \
        .one_way_direction("cluster-agent", "cluster-agent-container", type="encloses") \
        .one_way_direction("cluster-agent", "cluster-agent-cm", type="claims") \
        .one_way_direction("cluster-agent", "node", type="scheduled_on") \
        .one_way_direction("cluster-agent-container", "cluster-agent-cm", type="mounts") \
        .one_way_direction("cluster-agent", "cluster-agent-secret", type="uses_value") \
        .one_way_direction("cluster-agent-container", "cluster-agent-cluster-agent", type="runs")

    query_and_assert(cliv1, cluster_name, namespace, expected_topology)


def test_node_agent_topology(ansible_var, cliv1):
    k8s_node_count = int(ansible_var("agent_k8s_size"))
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")

    node_agent = "stackstate-cluster-agent-agent"
    cluster_agent = "stackstate-cluster-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node-agent-svc", type="service", name=node_agent) \
        .component("node-agent-ds", type="daemonset", name=node_agent) \
        .component("node-agent-cm", type="configmap", name=node_agent) \
        .component("cluster-agent-secret", type="secret", name=cluster_agent) \
        .one_way_direction("namespace", "node-agent-svc", type="encloses") \
        .one_way_direction("namespace", "node-agent-ds", type="encloses") \
        .repeated(
            k8s_node_count,
            lambda matcher: matcher
            .component("node", type="node", name=r"ip-.*")
            .component("node-agent", type="pod", name=fr"{node_agent}-.*")
            .component("node-agent-main-container", type="container", name="agent")
            .component("node-agent-process-container", type="container", name="process-agent")
            .component("cgroups-vol", type="volume", name="cgroups")
            .component("node-agent-process-agent", type="stackstate-agent", name="process-agent")
            .component("node-agent-trace-agent", type="stackstate-agent", name="trace-agent")
            .component("node-agent-main-agent", type="stackstate-agent", name="agent run")
            .one_way_direction("node-agent-ds", "node-agent", type="controls")
            .one_way_direction("node-agent-svc", "node-agent", type="exposes")
            .one_way_direction("node-agent", "node", type="scheduled_on")
            .one_way_direction("node-agent", "node-agent-main-container", type="encloses")
            .one_way_direction("node-agent", "node-agent-process-container", type="encloses")
            .one_way_direction("node-agent", "node-agent-cm", type="claims")
            .one_way_direction("node-agent-main-container", "node-agent-cm", type="mounts")
            .one_way_direction("node-agent-process-container", "node-agent-cm", type="mounts")
            .one_way_direction("node-agent-main-container", "cgroups-vol", type="mounts")
            .one_way_direction("node-agent-process-container", "cgroups-vol", type="mounts")
            .one_way_direction("node-agent", "cluster-agent-secret", type="uses_value")
            .one_way_direction("node-agent-process-container", "node-agent-process-agent", type="runs")
            .one_way_direction("node-agent-main-container", "node-agent-trace-agent", type="runs")
            .one_way_direction("node-agent-main-container", "node-agent-main-agent", type="runs")
        )

    matched_res = query_and_assert(cliv1, cluster_name, namespace, expected_topology)

    node_agent_pod_name = matched_res.component(("node-agent", 0)).name
    assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-main-container", 0)).tags
    assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-process-container", 0)).tags
    assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-process-agent", 0)).tags
    assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-trace-agent", 0)).tags
    assert f"pod-name:{node_agent_pod_name}" in matched_res.component(("node-agent-main-agent", 0)).tags


def test_checks_agent_topology(ansible_var, cliv1):
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("monitoring_namespace")

    checks_agent = "stackstate-cluster-agent-clusterchecks"
    cluster_agent = "stackstate-cluster-agent"

    expected_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("node", type="node", name=r"ip-.*") \
        .component("checks-agent-deployment", type="deployment", name=checks_agent) \
        .component("checks-agent-rs", type="replicaset", name=fr"{checks_agent}-.*") \
        .component("checks-agent", type="pod", name=fr"{checks_agent}-.*-.*") \
        .component("checks-agent-container", type="container", name="cluster-agent") \
        .component("cluster-agent-secret", type="secret", name=cluster_agent) \
        .component("cluster-agent-main-agent", type="stackstate-agent", name="agent run") \
        .one_way_direction("namespace", "checks-agent-deployment", type="encloses") \
        .one_way_direction("checks-agent-deployment", "checks-agent-rs", type="controls") \
        .one_way_direction("checks-agent-rs", "checks-agent", type="controls") \
        .one_way_direction("checks-agent", "node", type="scheduled_on") \
        .one_way_direction("checks-agent", "checks-agent-container", type="encloses") \
        .one_way_direction("checks-agent", "cluster-agent-secret", type="uses_value") \
        .one_way_direction("checks-agent-container", "cluster-agent-main-agent", type="runs")

    query_and_assert(cliv1, cluster_name, namespace, expected_topology)


def query_and_assert(cliv1, cluster_name: str, namespace: str, expected_topology: TopologyMatcher):
    current_agent_topology = cliv1.topology(
        f"(label IN ('cluster-name:{cluster_name}') AND label IN ('namespace:{namespace}'))"
        f" OR (type IN ('node', 'namespace'))")
    possible_matches = expected_topology.find(current_agent_topology)
    return possible_matches.assert_exact_match()
