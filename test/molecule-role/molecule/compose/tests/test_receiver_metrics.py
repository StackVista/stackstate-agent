import json
import os
import util
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('trace-java-demo-compose')


def test_container_metrics(host):
    url = "http://localhost:7070/api/topic/sts_multi_metrics?limit=3000"

    def wait_for_metrics():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-sts-multi-metrics.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        def get_keys(m_host):
            return next(set(message["message"]["MultiMetric"]["values"].keys())
                        for message in json_data["messages"]
                        if message["message"]["MultiMetric"]["host"] == m_host and
                           "containerId" in message["message"]["MultiMetric"]["tags"]

                        )

        expected = {"netRcvdPs", "memCache", "totalPct", "wbps",
                    "systemPct", "rbps", "memRss", "netSentBps", "netSentPs", "netRcvdBps", "userPct"}
        assert get_keys("trace-java-demo") == expected

    util.wait_until(wait_for_metrics, 180, 3)


def test_no_datadog_metrics(host):
    url = "http://localhost:7070/api/topic/sts_multi_metrics?limit=3000"

    def wait_for_metrics():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        with open("./topic-sts-multi-metrics.json", 'w') as f:
            json.dump(json_data, f, indent=4)

        metrics = {}
        for message in json_data["messages"]:
            for m_name in message["message"]["MultiMetric"]["values"].keys():
                if m_name not in metrics:
                    metrics[m_name] = []

                values = [message["message"]["MultiMetric"]["values"][m_name]]
                metrics[m_name] += values

        # assert that we don't see any datadog metrics
        # datadog.dogstatsd is a part of external dependency datadog-go
        datadog_metrics = [(key, value) for key, value in metrics.items() if key.startswith("datadog") and not key.startswith("datadog.dogstatsd") ]
        assert len(datadog_metrics) == 0, 'datadog metrics found in sts_multi_metrics: [%s]' % ', '.join(map(str, datadog_metrics))

    util.wait_until(wait_for_metrics, 60, 3)
