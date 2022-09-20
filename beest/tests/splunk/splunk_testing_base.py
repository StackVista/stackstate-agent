import logging
import testinfra.utils.ansible_runner
import requests
import random

from ststest import TopologyMatcher
from urllib3.exceptions import InsecureRequestWarning


class SplunkBase:
    def __init__(self, host, ansible_var, log, yard_location):
        self.host = host
        self.log: logging.Logger = log
        self.yard_location: str = yard_location
        self.ansible_var = ansible_var

        self.yard_id = ansible_var("yard_id")
        self.splunk_instance = self.set_splunk_instance(yard_location)

        # Authentication Details
        splunk_user = ansible_var("splunk_user")
        splunk_pass = ansible_var("splunk_pass")

        # Types of Splunk Integrations
        self.topology = SplunkTopologyBase(ansible_var, self.splunk_instance, splunk_user, splunk_pass)
        self.event = SplunkEventBase(ansible_var, self.splunk_instance, splunk_user, splunk_pass)
        self.health = SplunkHealthBase(ansible_var, self.splunk_instance, splunk_user, splunk_pass)
        self.metric = SplunkMetricBase(ansible_var, self.splunk_instance, splunk_user, splunk_pass)

    # Force the query to be in the Splunk StackPack with the yard id as the instance
    def stackpack_topology_query(self, query_suffix: str):
        query_prefix = f"label IN ('stackpack:splunk', 'splunk-instance:{self.yard_id}')"
        return f"{query_prefix} AND {query_suffix}"

    # We are selecting local in the testinfra_hosts, This will expose all the local ansible variables
    # Only problem is that given local the ansible_host will be localhost and not the dynamic splunk instance ip
    # For this we can load up the inventory again and select another host, we can then temp retrieve variables from this
    # host like the splunk instance
    @staticmethod
    def set_splunk_instance(yard_location) -> str:
        # Open up the ansible_inventory inventory again based on the same one we created the testinfra_hosts with
        splunk_ansible_inventory = testinfra.utils.ansible_runner.AnsibleRunner(f'{yard_location}/ansible_inventory')

        # Now we select the other host, not local
        splunk_variables = splunk_ansible_inventory.get_variables("splunk")

        # From this host we can extract a few common variables but the important one is the ansible_host variable.
        # The ansible_host will contain the actual IP of the Splunk Machine and not Localhost
        splunk_protocol = splunk_variables.get("splunk_instance_protocol")
        splunk_host = splunk_variables.get("ansible_host")
        splunk_port = splunk_variables.get("splunk_instance_port")

        # Combine the results in a valid URL that we can query splunk with
        splunk_instance = "{}://{}:{}".format(splunk_protocol, splunk_host, splunk_port)

        return splunk_instance


class SplunkCommonBase:
    def __init__(self, ansible_var, splunk_instance, splunk_user, splunk_pass):
        # The calculated splunk instance
        self.splunk_instance = splunk_instance

        # Create a session to control all requests
        self.session = requests.Session()
        self.session.verify = False

        # Authentication Details
        self.splunk_user = splunk_user
        self.splunk_pass = splunk_pass

        # Disable Security Warning
        requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

    def post_to_services_receivers_simple(self, json_data, param_data=None):
        self.session.post("%s/services/receivers/simple" % self.splunk_instance,
                          json=json_data,
                          params=param_data,
                          auth=(self.splunk_user, self.splunk_pass)) \
            .raise_for_status()


class SplunkTopologyBase(SplunkCommonBase):
    def __init__(self, ansible_var, splunk_instance, splunk_user, splunk_pass):
        super().__init__(ansible_var, splunk_instance, splunk_user, splunk_pass)

    def publish_random_server_component(self) -> str:
        component_id = "server_{}".format(random.randint(0, 10000))

        self._post_component(component_id=component_id,
                             component_type="server",
                             description="Topology Server Component")

        return component_id

    def _post_component(self, component_id: str, component_type: str,
                        description: str = "Component Description",
                        topo_type: str = "component"):
        json_data = {
            "topo_type": topo_type,
            "id": component_id,
            "type": component_type,
            "description": description
        }
        self.post_to_services_receivers_simple(json_data=json_data)

    def _post_relation(self, relation_type: str, source_id: str, target_id: str,
                       description: str = "Relation Description",
                       topo_type: str = "relation"):
        json_data = {
            "topo_type": topo_type,
            "type": relation_type,
            "sourceId": source_id,
            "targetId": target_id,
            "description": description
        }
        self.post_to_services_receivers_simple(json_data=json_data)


class SplunkHealthBase(SplunkCommonBase):
    def __init__(self, ansible_var, splunk_instance, splunk_user, splunk_pass):
        super().__init__(ansible_var, splunk_instance, splunk_user, splunk_pass)

    # Core method for posting events to Splunk
    def _post_health(self, check_state_id: str, name: str, status: str, topology_element_identifier: str,
                     message: str = None):

        json_data = {
            "check_state_id": check_state_id,
            "name": name,
            "health": status,
            "topology_element_identifier": topology_element_identifier
            }

        if message is not None:
            json_data["message"] = message

        self.post_to_services_receivers_simple(json_data=json_data)


class SplunkEventBase(SplunkCommonBase):
    def __init__(self, ansible_var, splunk_instance, splunk_user, splunk_pass):
        super().__init__(ansible_var, splunk_instance, splunk_user, splunk_pass)

    # Core method for posting events to Splunk
    def _post_event(self, status: str, host: str, source_type: str,
                    description: str = "Event Description"):
        param_data = {
            "host": host,
            "sourcetype": source_type
        }

        json_data = {
            "status": status,
            "description": description
        }

        self.post_to_services_receivers_simple(json_data=json_data, param_data=param_data)


class SplunkMetricBase(SplunkCommonBase):
    def __init__(self, ansible_var, splunk_instance, splunk_user, splunk_pass):
        super().__init__(ansible_var, splunk_instance, splunk_user, splunk_pass)

    # Core method for posting events to Splunk
    def _post_metric(self, topo_type: str, identifier: str, value: int, qa: str, host: str, source_type: str):
        param_data = {
            "host": host,
            "sourcetype": source_type
        }

        json_data = {
            "topo_type": topo_type,
            "metric": identifier,
            "value": value,
            "qa": qa
        }

        self.post_to_services_receivers_simple(json_data=json_data, param_data=param_data)
