import json
import util
from conftest import STS_CONTEXT_FILE

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def test_agents_running(cliv1):
    def wait_for_metrics():
        expected_metrics = ["stackstate_agent_running", "stackstate_cluster_agent_running"]
        for expected_metric in expected_metrics:
            json_data = cliv1.promql_script(f'Telemetry.instantPromql(\\\"{expected_metric}\\\")', expected_metric)
            for result in json_data["result"]:
                if result["_type"] == "MetricTimeSeriesResult":
                    timeseries = result["timeSeries"]
                    points = timeseries["points"]
                    for point in points:
                        assert point[1] == 1.0

    util.wait_until(wait_for_metrics, 60, 3)


def test_container_metrics(cliv1):
    def wait_for_metrics():
        expected_metrics = ["netRcvdPs", "memCache", "totalPct", "wbps", "systemPct", "rbps", "memRss", "netSentBps",
                            "netSentPs", "netRcvdBps", "userPct"]
        non_zeros = ["memRss", "systemPct"]
        non_zeros_result = [False, False]

        for expected_metric in expected_metrics:
            json_data = cliv1.promql_script(f'Telemetry.instantPromql(\\\"{expected_metric}\\\")', expected_metric)
            for result in json_data["result"]:
                if result["_type"] == "MetricTimeSeriesResult":
                    query = result["query"]
                    assert query["query"] == expected_metric
                    if expected_metric in non_zeros:
                        timeseries = result["timeSeries"]
                        points = timeseries["points"]
                        for point in points:
                            if point[1] > 0:
                                non_zeros_result[non_zeros.index(expected_metric)] = True

        assert all(non_zeros_result)

    util.wait_until(wait_for_metrics, 60, 3)


# TODO: HTTP Metrics has been updated to use a new topic etc.
# TODO: - pod_http_requests_count
# TODO: - pod_http_response_time_seconds_bucket

#  def test_agent_http_metrics(cliv1):
#      def wait_for_metrics():
#          json_data = cliv1.topic_api("sts_multi_metrics", config_location=STS_CONTEXT_FILE)
#
#          def get_keys():
#              return next(set(message["message"]["MultiMetric"]["values"].keys())
#                          for message in json_data["messages"]
#                          if message["message"]["MultiMetric"]["name"] == "connection metric" and
#                          "code" in message["message"]["MultiMetric"]["tags"] and
#                          message["message"]["MultiMetric"]["tags"]["code"] == "2xx"
#                          )
#
#          expected = {"http_requests_per_second", "http_response_time_seconds"}
#
#          assert get_keys().pop() in expected
#
#      util.wait_until(wait_for_metrics, 30, 3)

def _check_metric_exists(json_data, root_tag):
    if root_tag in json_data:
        data = json_data[root_tag]
        if len(data) >= 1:
            return True
    return False


def _check_contains_tag(json_data, root_tag, searched_tag):
    if _check_metric_exists(json_data, root_tag):
        data = json_data[root_tag]
        for result in data:
            timeseries = result["timeSeries"]
            _id = timeseries["id"]
            groups = _id["groups"]
            if searched_tag in groups:
                return True
    return False


def test_agent_kubernetes_metrics(cliv1):
    def wait_for_metrics():
        docker_metric_exists = False
        k8s_metric_exists = False
        docker_contains_kcn = False
        k8s_contains_kcn = False
        docker_expected_metrics = ["docker_containers_running", "docker_containers_scheduled"]
        kubernetes_expected_metrics = ["kubernetes_state_pod_ready", "kubernetes_state_pod_scheduled"]

        for docker_expected_metric in docker_expected_metrics:
            json_data_docker = cliv1.promql_script(f'Telemetry.instantPromql\\(\\\"{docker_expected_metric}\\\"\\)',
                                                   docker_expected_metric)
            docker_metric_exists = _check_metric_exists(json_data_docker, "result")
            docker_contains_kcn = _check_contains_tag(json_data_docker, "result", "kube_cluster_name")
            docker_contains_cn = _check_contains_tag(json_data_docker, "result", "cluster_name")

            docker_kcn = docker_contains_kcn or docker_contains_cn

            if docker_metric_exists or docker_contains_kcn:
                break

        for kubernetes_expected_metric in kubernetes_expected_metrics:
            json_data_kubernetes = cliv1.promql_script(
                f'Telemetry.instantPromql\\(\\\"{kubernetes_expected_metric}\\\"\\)', kubernetes_expected_metric)
            k8s_metric_exists = _check_metric_exists(json_data_kubernetes, "result")
            k8s_contains_kcn = _check_contains_tag(json_data_kubernetes, "result", "kube_cluster_name")
            k8s_contains_cn = _check_contains_tag(json_data_kubernetes, "result", "cluster_name")

            k8s_kcn = k8s_contains_kcn or k8s_contains_cn

            if k8s_metric_exists or k8s_contains_kcn:
                break

        final_decision = (docker_kcn or k8s_kcn) and (docker_metric_exists or k8s_metric_exists)

        assert final_decision, 'No kubernetes metrics found'

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_kubelet_metrics(cliv1):
    def wait_for_metrics():
        expected_metrics = ["kubernetes_kubelet_volume_stats_available_bytes",
                            "kubernetes_kubelet_volume_stats_used_bytes"]
        metric_exists = False
        contains_namespace = False

        for expected_metric in expected_metrics:
            json_data = cliv1.promql_script(f'Telemetry.instantPromql\\(\\\"{expected_metric}\\\"\\)',
                                            expected_metric)
            metric_exists = _check_metric_exists(json_data, "result") or metric_exists
            contains_namespace = _check_contains_tag(json_data, "result", "namespace") or contains_namespace

        final_decision = metric_exists and contains_namespace
        assert final_decision, 'No kubelet metrics found'

    util.wait_until(wait_for_metrics, 60, 3)
