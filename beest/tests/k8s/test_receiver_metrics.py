import json
import util
from conftest import STS_CONTEXT_FILE

testinfra_hosts = [f"ansible://local?ansible_inventory=../../sut/yards/k8s/ansible_inventory"]


def test_agents_running(cliv1):
    def wait_for_metrics():
        data_points = ["stackstate_agent_running", "stackstate_cluster_agent_running"]
        for data_point in data_points:
            json_data = cliv1.promql_script(f'Telemetry.instantPromql\(\\\"{data_point}\\\"\)', data_point)
            for result in json_data["result"]:
                if result["_type"] == "MetricTimeSeriesResult":
                    timeseries = result["timeSeries"]
                    points = timeseries["points"]
                    for point in points:
                        assert point[1] == 1.0

    util.wait_until(wait_for_metrics, 60, 3)


def test_container_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", limit=4000, config_location=STS_CONTEXT_FILE)

        metrics = {}
        for message in json_data["messages"]:
            for m_name in message["message"]["MultiMetric"]["values"].keys():
                if m_name not in metrics:
                    metrics[m_name] = []

                values = [message["message"]["MultiMetric"]["values"][m_name]]
                metrics[m_name] += values

        expected = {"netRcvdPs", "memCache", "totalPct", "wbps",
                    "systemPct", "rbps", "memRss", "netSentBps", "netSentPs", "netRcvdBps", "userPct"}
        for e in expected:
            assert e in metrics, "%s metric was not found".format(e)

        check_non_zero("memRss", metrics)
        check_non_zero("systemPct", metrics)

    util.wait_until(wait_for_metrics, 60, 3)


def check_non_zero(metric, metrics):
    for v in metrics[metric]:
        if v > 0:
            return
    assert False, "all '%s' metric are '0'".format(metric)


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

def test_agent_kubernetes_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", config_location=STS_CONTEXT_FILE)

        def contains_key():
            for message in json_data["messages"]:
                if (message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                    "kube_cluster_name" in message["message"]["MultiMetric"]["tags"] and
                    (
                        (
                            "docker.containers.running" in message["message"]["MultiMetric"]["values"].keys() or
                            "docker.containers.scheduled" in message["message"]["MultiMetric"]["values"].keys()
                        ) or
                        (
                            "kubernetes_state.pod.ready" in message["message"]["MultiMetric"]["values"].keys() or
                            "kubernetes_state.pod.scheduled" in message["message"]["MultiMetric"]["values"].keys()
                        ))):
                    return True
            return False

        assert contains_key(), 'No kubernetes metrics found'

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_kubernetes_state_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", config_location=STS_CONTEXT_FILE)

        def contains_key():
            for message in json_data["messages"]:
                if (message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                   "kube_cluster_name" in message["message"]["MultiMetric"]["tags"] and
                    (
                        (
                            "docker.containers.running" in message["message"]["MultiMetric"]["values"] or
                            "docker.containers.scheduled" in message["message"]["MultiMetric"]["values"]
                        ) or
                        (
                            "kubernetes_state.pod.ready" in message["message"]["MultiMetric"]["values"] or
                            "kubernetes_state.pod.scheduled" in message["message"]["MultiMetric"]["values"]
                        ))):
                    return True
            return False

        assert contains_key(), 'No kubernetes_state metrics found'

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_kubelet_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", limit=3000, config_location=STS_CONTEXT_FILE)

        def contains_key():
            for message in json_data["messages"]:
                if (message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                    "namespace" in message["message"]["MultiMetric"]["tags"] and
                    ("kubernetes.kubelet.volume.stats.available_bytes" in message["message"]["MultiMetric"]["values"] or
                     "kubernetes.kubelet.volume.stats.used_bytes" in message["message"]["MultiMetric"]["values"])):
                    return True
            return False

        assert contains_key(), 'No kubelet metrics found'

    util.wait_until(wait_for_metrics, 60, 3)
