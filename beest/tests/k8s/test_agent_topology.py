from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_agent_base_topology(ansible_var, cliv1):
    cluster_name = ansible_var("cluster_name")
    namespace = ansible_var("namespace")

    cluster_agent = "stackstate-cluster-agent"
    node_agent = "stackstate-cluster-agent-agent"
    checks_agent = "stackstate-cluster-agent-clusterchecks"

    # TODO:
    # match on component labels

    NODE_COUNT = 2

    expected_agent_topology = \
        TopologyMatcher() \
            .component("cluster-agent-deployment", type="deployment", name=cluster_agent) \
            .component("cluster-agent-rs", type="replicaset", name=fr"{cluster_agent}-\w{{9,10}}") \
            .component("cluster-agent", type="pod", name=fr"{cluster_agent}-\w{{9,10}}-\w{{5}}") \
            .one_way_direction("cluster-agent-deployment", "cluster-agent-rs", type="controls") \
            .one_way_direction("cluster-agent-rs", "cluster-agent", type="controls") \
            .component("checks-agent-deployment", type="deployment", name=checks_agent) \
            .component("checks-agent-rs", type="replicaset", name=fr"{checks_agent}-.*") \
            .component("checks-agent", type="pod", name=fr"{checks_agent}-.*-.*") \
            .one_way_direction("checks-agent-deployment", "checks-agent-rs", type="controls") \
            .one_way_direction("checks-agent-rs", "checks-agent", type="controls") \
            .component("node-agent-daemonset", type="daemonset", name=node_agent) \
            .repeated(
                NODE_COUNT,
                lambda matcher: matcher
                .component("node-agent", type="pod", name=fr"{node_agent}-.*")
                .one_way_direction("node-agent-daemonset", "node-agent", type="controls")
            )
            # .component("namespace", type="namespace", name=namespace) \
            # .component("node1", type="node") \
            # .component("node2", type="node") \
            # .component("cluster-agent-container", type="container", name=fr"{cluster_agent}-.*") \
            # .one_way_direction("cluster-agent", "node1", type="scheduled_on") \
            # .component("checks-agent-container", type="container", name=r"stackstate-cluster-agent-clusterchecks") \
            # .component("node-agent1", type="pod", name=r"stackstate-cluster-agent-agent-.*") \
            # .component("node-agent2", type="pod", name=r"stackstate-cluster-agent-agent-.*") \
            # .component("node-agent-service", type="service", name=r"stackstate-cluster-agent-agent") \
            # .component("node-agent-ds", type="daemonset", name=r"stackstate-cluster-agent-agent") \
            # .one_way_direction("node-agent1", "node1", type="scheduled_on") \
            # .one_way_direction("node-agent2", "node2", type="scheduled_on") \
            # .one_way_direction("node-agent-service", "node-agent1", type="exposes") \
            # .one_way_direction("node-agent-service", "node-agent2", type="exposes") \
            # .one_way_direction("node-agent-ds", "node-agent1", type="controls") \
            # .one_way_direction("node-agent-ds", "node-agent2", type="controls") \

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
