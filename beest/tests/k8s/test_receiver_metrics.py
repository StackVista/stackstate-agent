import json
import util

testinfra_hosts = ["local"]


def test_agents_running(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics")

        metrics = {}
        for message in json_data["messages"]:
            for m_name in message["message"]["MultiMetric"]["values"].keys():
                if m_name not in metrics:
                    metrics[m_name] = []

                values = [message["message"]["MultiMetric"]["values"][m_name]]
                metrics[m_name] += values

        for v in metrics["stackstate.agent.running"]:
            assert v == 1.0
        for v in metrics["stackstate.cluster_agent.running"]:
            assert v == 1.0

        # assert that we don't see any datadog metrics
        datadog_metrics = [(key, value) for key, value in metrics.items() if key.startswith("datadog")]
        assert len(datadog_metrics) == 0, 'datadog metrics found in sts_multi_metrics: [%s]' % ', '.join(map(str, datadog_metrics))

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_http_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics")

        def get_keys():
            return next(set(message["message"]["MultiMetric"]["values"].keys())
                        for message in json_data["messages"]
                        if message["message"]["MultiMetric"]["name"] == "connection metric" and
                        "code" in message["message"]["MultiMetric"]["tags"] and
                        message["message"]["MultiMetric"]["tags"]["code"] == "any"
                        )

        expected = {"http_requests_per_second", "http_response_time_seconds"}

        assert get_keys().pop() in expected

    util.wait_until(wait_for_metrics, 30, 3)


def test_agent_kubernetes_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics")

        def contains_key():
            for message in json_data["messages"]:
                if (message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                    "cluster_name" in message["message"]["MultiMetric"]["tags"] and
                    ("kubernetes_state.container.running" in message["message"]["MultiMetric"]["values"].keys() or
                     "kubernetes_state.pod.scheduled" in message["message"]["MultiMetric"]["values"].keys())):
                    return True
            return False

        assert contains_key(), 'No kubernetes metrics found'

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_kubernetes_state_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics")

        def contains_key():
            for message in json_data["messages"]:
                if (message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                    "cluster_name" in message["message"]["MultiMetric"]["tags"] and
                    ("kubernetes_state.container.running" in message["message"]["MultiMetric"]["values"] or
                     "kubernetes_state.pod.scheduled" in message["message"]["MultiMetric"]["values"])):
                    return True
            return False

        assert contains_key(), 'No kubernetes_state metrics found'

    util.wait_until(wait_for_metrics, 60, 3)


def test_agent_kubelet_metrics(cliv1):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", limit=3000)

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
