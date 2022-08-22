from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_agent_base_topology(ansible_var, cliv1):
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("namespace")

    cluster_agent = "stackstate-cluster-agent"
    node_agent = "stackstate-cluster-agent-agent"
    checks_agent = "stackstate-cluster-agent-clusterchecks"

    # TODO:
    # match on component labels

    NODE_COUNT = 2

    expected_agent_topology = \
        TopologyMatcher() \
            .component("cluster-agent-svc", type="service", name=cluster_agent) \
            .component("cluster-agent-deployment", type="deployment", name=cluster_agent) \
            .component("cluster-agent-rs", type="replicaset", name=fr"{cluster_agent}-\w{{9,10}}") \
            .component("cluster-agent", type="pod", name=fr"{cluster_agent}-\w{{9,10}}-\w{{5}}") \
            .component("cluster-agent-cm", type="configmap", name=cluster_agent) \
            .component("cluster-agent-secret", type="secret", name=cluster_agent) \
            .one_way_direction("cluster-agent-deployment", "cluster-agent-rs", type="controls") \
            .one_way_direction("cluster-agent-rs", "cluster-agent", type="controls") \
            .one_way_direction("cluster-agent-svc", "cluster-agent", type="exposes") \
            .one_way_direction("cluster-agent", "cluster-agent-cm", type="claims") \
            .one_way_direction("cluster-agent", "cluster-agent-secret", type="uses_value") \
            .component("checks-agent-deployment", type="deployment", name=checks_agent) \
            .component("checks-agent-rs", type="replicaset", name=fr"{checks_agent}-.*") \
            .component("checks-agent", type="pod", name=fr"{checks_agent}-.*-.*") \
            .one_way_direction("checks-agent-deployment", "checks-agent-rs", type="controls") \
            .one_way_direction("checks-agent-rs", "checks-agent", type="controls") \
            .one_way_direction("checks-agent", "cluster-agent-secret", type="uses_value") \
            .component("node-agent-svc", type="service", name=node_agent) \
            .component("node-agent-ds", type="daemonset", name=node_agent) \
            .component("node-agent-cm", type="configmap", name=node_agent) \
            .repeated(
                NODE_COUNT,
                lambda matcher: matcher
                    .component("node", type="node", name=r"node-.*")
                    .component("node-agent", type="pod", name=fr"{node_agent}-.*")
                    .one_way_direction("node-agent-ds", "node-agent", type="controls")
                    .one_way_direction("node-agent-svc", "node-agent", type="exposes")
                    .one_way_direction("node-agent", "node", type="scheduled_on")
                    .one_way_direction("node-agent", "node-agent-cm", type="claims")
                    .one_way_direction("node-agent", "cluster-agent-secret", type="uses_value")
            ) \
            .one_way_direction("cluster-agent", ("node", 0), type="scheduled_on") \
            .one_way_direction("checks-agent", ("node", 0), type="scheduled_on") \
            .component("namespace", type="namespace", name=namespace) \


    current_agent_topology = cliv1.topology(
        f"(label IN ('namespace:{namespace}') and label in ('app.kubernetes.io/name:cluster-agent'))" +
        " or (type in ('node', 'namespace'))"
    )
    possible_matches = expected_agent_topology.find(current_agent_topology)
    match_result = possible_matches.assert_exact_match()

    # assert match_result.component("cluster-agent-container").attributes["tags"] is [
    #     "pod-name:%s" % stackstate_cluster_agent_container_pod,
    #     "namespace:%s" % namespace,
    #     "cluster-name:%s" % cluster_name
    # ]
