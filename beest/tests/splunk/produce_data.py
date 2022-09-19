import requests

from urllib3.exceptions import InsecureRequestWarning


def publish_components(ansible_var, splunk_instance, component_id):
    # Disable Security Warning
    requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

    # Create a session to control all requests
    session = requests.Session()
    session.verify = False

    splunk_user = ansible_var("splunk_user")
    splunk_pass = ansible_var("splunk_pass")

    json_data = {"topo_type": "component", "id": component_id, "type": "server", "description": "My important server 1"}
    session.post("%s/services/receivers/simple" % splunk_instance, json=json_data, auth=(splunk_user, splunk_pass))\
        .raise_for_status()
