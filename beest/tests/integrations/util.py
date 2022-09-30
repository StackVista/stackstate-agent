import time
import json


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


def match_partial_event(event, json_data):
    for message in json_data["messages"]:
        p = message["message"]
        if "GenericEvent" in p:
            _data = p["GenericEvent"]
            if _data == dict(_data, **event):
                return True

    return False


def assert_metrics_check_instance(host, test_data, expected_metrics, check_instance):
    hostname = host.ansible.get_variables()["inventory_hostname"]
    print(hostname)

    def wait_for_metrics():
        def get_keys(m_host):
            return set(
                ''.join(message["message"]["MultiMetric"]["values"].keys())
                for message in test_data["messages"]
                if message["message"]["MultiMetric"]["name"] == "convertedMetric" and
                message["message"]["MultiMetric"]["host"] == m_host and
                check_instance in message["message"]["MultiMetric"]["labels"]
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
