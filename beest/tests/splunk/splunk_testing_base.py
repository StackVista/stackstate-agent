import logging
import testinfra.utils.ansible_runner
import requests
import random

from ststest import TopologyMatcher
from urllib3.exceptions import InsecureRequestWarning


class SplunkTestingBase:
    def __init__(self, host, ansible_var, log, yard_location):
        self.host = host
        self.log: logging.Logger = log
        self.yard_location: str = yard_location
        self.ansible_var = ansible_var

        self.yard_id = ansible_var("yard_id")
        self.splunk_instance = self.set_splunk_instance(yard_location)
        self.topology = SplunkTestingTopologyBase(ansible_var, self.splunk_instance)

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


class SplunkTestingTopologyBase:
    def __init__(self, ansible_var, splunk_instance):
        # The calaculated splunk instance
        self.splunk_instance = splunk_instance

        # Create a session to control all requests
        self.session = requests.Session()
        self.session.verify = False

        # Authentication Details
        self.splunk_user = ansible_var("splunk_user")
        self.splunk_pass = ansible_var("splunk_pass")

        # Disable Security Warning
        requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

    # Posts a component to splunk with the type of server
    def publish_random_server_component(self) -> str:
        component_id = "server_{}".format(random.randint(0, 10000))

        self._post_component(component_id=component_id,
                             component_type="server",
                             component_description="Topology Server Component",
                             component_topo_type="component")

        return component_id

    # Core method for posting components to Splunk
    def _post_component(self,
                        component_id: str,
                        component_type: str,
                        component_description: str,
                        component_topo_type: str):
        json_data = {
            "topo_type": component_topo_type,
            "id": component_id,
            "type": component_type,
            "description": component_description
        }
        self.session.post("%s/services/receivers/simple" % self.splunk_instance,
                          json=json_data,
                          auth=(self.splunk_user, self.splunk_pass)) \
            .raise_for_status()
