import util
import pytest

from conftest import STS_CONTEXT_FILE
from ststest import TopologyMatcher

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


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
        .repeated(
            k8s_node_count,
            lambda matcher: matcher
            .component("mehdb-pod", type="pod", name=r"mehdb-.*")
            .component("mehdb-container", type="container", name="shard")
            .component("mehdb-pv", type="persistent-volume", name=r"pvc-.*")
            .component("mehdb-pvc", type="persistent-volume-claim", name=r"data-mehdb.*")
            .one_way_direction("mehdb-ss", "mehdb-pod", type="controls")
            .one_way_direction("mehdb-svc", "mehdb-pod", type="exposes")
            .one_way_direction("mehdb-pod", "mehdb-pvc", type="claims")
            .one_way_direction("mehdb-pod", "mehdb-container", type="encloses")
            .one_way_direction("mehdb-container", "mehdb-pvc", type="mounts")
            .one_way_direction("mehdb-pvc", "mehdb-pv", type="exposes")
        ) \
        .component("example-ingress", type="ingress", name="example-ingress") \
        .component("apple-svc", type="service", name="apple-service") \
        .component("banana-svc", type="service", name="banana-service") \
        .one_way_direction("example-ingress", "apple-svc", type="routes") \
        .one_way_direction("example-ingress", "banana-svc", type="routes") \

    current_agent_topology = cliv1.topology(
        f"(label IN ('cluster-name:{cluster_name}') AND label IN ('namespace:{namespace}'))"
        f" OR type IN ('namespace', 'persistent-volume', 'persistent-volume-claim')",
        alias="workload", config_location=STS_CONTEXT_FILE)
    possible_matches = expected_agent_topology.find(current_agent_topology)
    possible_matches.assert_exact_match()

