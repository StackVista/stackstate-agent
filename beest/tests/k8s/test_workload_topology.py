import util
from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_workload_topology(ansible_var, cliv1):
    k8s_node_count = int(ansible_var("agent_k8s_size"))
    cluster_name = ansible_var("agent_cluster_name")
    namespace = ansible_var("test_namespace")

    # cronjob creates a job every minute, by default k8s retains 3 finished jobs
    successfulJobsHistoryLimit = 3

    expected_agent_topology = TopologyMatcher() \
        .component("namespace", type="namespace", name=namespace) \
        .component("pod-svc", type="service", name="pod-service") \
        .component("server-pod", type="pod", name="pod-server") \
        .one_way_direction("pod-svc", "server-pod", type="exposes") \
        .component("google-svc", type="service", name="google-service") \
        .component("hello-cj", type="cronjob", name="hello") \
        .repeated(
            successfulJobsHistoryLimit,
            lambda matcher: matcher
            .component("hello-j", type="job", name=r"hello-.*")
            .one_way_direction("hello-cj", "hello-j", type="creates")
        ) \
        .component("countdown-j", type="job", name="countdown") \
        .component("countdown-pod", type="pod", name=r"countdown-.*") \
        .one_way_direction("countdown-j", "countdown-pod", type="controls") \
        .component("mehdb-ss", type="statefulset", name="mehdb") \
        .component("mehdb-svc", type="service", name="mehdb") \
        .one_way_direction("namespace", "mehdb-ss", type="encloses") \
        .one_way_direction("namespace", "mehdb-svc", type="encloses") \
        .repeated(
            k8s_node_count,
            lambda matcher: matcher
            .component("mehdb-pod", type="pod", name=r"mehdb-.*")
            .component("mehdb-container", type="container", name="shard")
            .component("mehdb-pvc", type="persistent-volume", name=r"pvc-.*")
            .one_way_direction("mehdb-ss", "mehdb-pod", type="controls")
            .one_way_direction("mehdb-svc", "mehdb-pod", type="exposes")
            .one_way_direction("mehdb-pod", "mehdb-pvc", type="claims")
            .one_way_direction("mehdb-pod", "mehdb-container", type="encloses")
            .one_way_direction("mehdb-container", "mehdb-pvc", type="mounts")
        ) \
        .component("example-ingress", type="ingress", name="example-ingress") \
        .component("apple-svc", type="service", name="apple-service") \
        .component("banana-svc", type="service", name="banana-service") \
        .one_way_direction("example-ingress", "apple-svc", type="routes") \
        .one_way_direction("example-ingress", "banana-svc", type="routes") \

    current_agent_topology = cliv1.topology(
        f"(label IN ('cluster-name:{cluster_name}') AND label IN ('namespace:{namespace}'))"
        f" OR type IN ('namespace', 'persistent-volume')", "workload")
    possible_matches = expected_agent_topology.find(current_agent_topology)
    possible_matches.assert_exact_match()


def test_container_runtime(ansible_var, cliv1):
    runtime = ansible_var("agent_k8s_runtime")
    if runtime == "dockerd":
        runtime = "docker"

    def wait_for_topology():
        topo = cliv1.topology("type = 'container'", "container-runtime")
        for c in topo.components:
            assert f"runtime:{runtime}" in c.tags

    util.wait_until(wait_for_topology(), 60, 3)
