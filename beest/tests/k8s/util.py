import time
import json

from conftest import STS_CONTEXT_FILE


def wait_until(someaction, timeout, period=0.25, *args, **kwargs):
    mustend = time.time() + timeout
    while True:
        try:
            someaction(*args, **kwargs)
            return
        except:
            if time.time() >= mustend:
                print("Waiting timed out after %d" % timeout)
                raise
            time.sleep(period)


def assert_topology_events(cliv1, test_name, topic, expected_topology_events):
    def wait_for_topology_events():
        json_data = cliv1.topic_api(topic, config_location=STS_CONTEXT_FILE)

        def _topology_event_data(event):
            for message in json_data["messages"]:
                p = message["message"]
                if "TopologyEvent" in p:
                    _data = p["TopologyEvent"]
                    if _data == dict(_data, **event):
                        return _data
            return None

        for t_e in expected_topology_events:
            print("Running assertion for: " + t_e["assertion"])
            assert _topology_event_data(t_e["event"]) is not None

    wait_until(wait_for_topology_events, 60, 3)


def assert_topology(cliv1, topic, expected_components):
    def assert_topology():
        json_data = cliv1.topic_api(topic, limit=1500, config_location=STS_CONTEXT_FILE)

        for c in expected_components:
            print("Running assertion for: " + c["assertion"])
            assert component_data(
                json_data=json_data,
                type_name=c["type"],
                external_id_assert_fn=c["external_id"],
                data_assert_fn=c["data"],
            ) is not None

    wait_until(assert_topology, 30, 3)


def assert_metrics(cliv1, hostname, expected_metrics):
    def wait_for_metrics():
        json_data = cliv1.topic_api("sts_multi_metrics", config_location=STS_CONTEXT_FILE)

        def get_keys(m_host):
            return set(
                ''.join(message["message"]["MultiMetric"]["values"].keys())
                for message in json_data["messages"]
                if message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                message["message"]["MultiMetric"]["host"] == m_host
            )

        assert all([expected_metric for expected_metric in expected_metrics if expected_metric in get_keys(hostname)])

    wait_until(wait_for_metrics, 30, 3)


def component_data(json_data, type_name, external_id_assert_fn, data_assert_fn):
    for message in json_data["messages"]:
        p = message["message"]["TopologyElement"]["payload"]
        if "TopologyComponent" in p and \
            p["TopologyComponent"]["typeName"] == type_name and \
            external_id_assert_fn(p["TopologyComponent"]["externalId"]):
            data = json.loads(p["TopologyComponent"]["data"])
            if data and data_assert_fn(data):
                return p["TopologyComponent"]["externalId"]
    return None


def event_data(event, json_data, hostname):
    for message in json_data["messages"]:
        p = message["message"]
        if "GenericEvent" in p and p["GenericEvent"]["host"] == hostname:
            _data = p["GenericEvent"]
            if _data == dict(_data, **event):
                return _data
    return None
