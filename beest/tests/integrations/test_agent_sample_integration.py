import util

testinfra_hosts = ["local"]


def kubernetes_event_data(event, json_data):
    for message in json_data["messages"]:
        p = message["message"]
        if "GenericEvent" in p:
            _data = p["GenericEvent"]
            if _data == dict(_data, **event):
                return _data
    return None


def test_sample_events(cliv1):

    def wait_for_events():
        json_data = cliv1.topic_api("sts_generic_events")

        service_event = {
            "name": "service-check.service-check",
            "title": "stackstate.agent.check_status",
            "eventType": "service-check",
            "tags": {
                "source_type_name": "service-check",
                "status": "OK",
                "check": "cpu"
            },
        }
        assert kubernetes_event_data(service_event, json_data) is not None

        http_event = {
            "name": "HTTP_TIMEOUT",
            "title": "URL timeout",
            "eventType": "HTTP_TIMEOUT",
            "tags": {
                "source_type_name": "HTTP_TIMEOUT"
            },
            "message": "Http request to http://localhost timed out after 5.0 seconds."
        }
        assert kubernetes_event_data(http_event, json_data) is not None

    util.wait_until(wait_for_events, 10, 5)
