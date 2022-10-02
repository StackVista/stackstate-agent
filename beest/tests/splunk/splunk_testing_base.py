from typing import Optional
from xml.etree.ElementTree import Element

import testinfra.utils.ansible_runner
import json
import logging
import requests
import random

from xml.etree import ElementTree
from pathlib import Path
from typing import TypedDict
from urllib3.exceptions import InsecureRequestWarning


class SplunkBase:
    # Splunk connection information for this base class
    # You will be able to connect to the splunk instance with the host and port
    splunk_instance: str
    splunk_protocol: str
    splunk_host: str
    splunk_port: str
    splunk_user: str
    splunk_pass: str

    def __init__(self, host, ansible_var, request, yard_location):
        # Store the ansible infra host in the base class
        self.host = host

        # The location where the yard is located for splunk
        self.yard_location: str = yard_location

        # The location where the yard is located for splunk
        self.request: str = request

        # A reference to the ansible_var variable selector
        self.ansible_var = ansible_var

        # Store the yard id for the splunk instance
        self.yard_id = ansible_var("yard_id")

        # Authentication Details
        self.splunk_user = ansible_var("splunk_user")
        self.splunk_pass = ansible_var("splunk_pass")

        # Set the splunk instance before using it
        # This will apply the protocol, host, port and instance
        self.set_splunk_instance(yard_location)

        # Types of Splunk Integrations
        self.topology = SplunkTopologyBase(ansible_var, request, self.splunk_instance, self.splunk_user,
                                           self.splunk_pass)
        self.event = SplunkEventBase(ansible_var, request, self.splunk_instance, self.splunk_user,
                                     self.splunk_pass)
        self.health = SplunkHealthBase(ansible_var, request, self.splunk_instance, self.splunk_user,
                                       self.splunk_pass)
        self.metric = SplunkMetricBase(ansible_var, request, self.splunk_instance, self.splunk_user,
                                       self.splunk_pass)

    # We are selecting local in the testinfra_hosts, This will expose all the local ansible variables
    # Only problem is that given local the ansible_host will be localhost and not the dynamic splunk instance ip
    # For this we can load up the inventory again and select another host, we can then temp retrieve variables from this
    # host like the splunk instance
    def set_splunk_instance(self, yard_location) -> None:
        # Open up the ansible_inventory inventory again based on the same one we created the testinfra_hosts with
        splunk_ansible_inventory = testinfra.utils.ansible_runner.AnsibleRunner(f'{yard_location}/ansible_inventory')

        # Now we select the other host, not local
        splunk_variables = splunk_ansible_inventory.get_variables("splunk")

        # From this host we can extract a few common variables but the important one is the ansible_host variable.
        # The ansible_host will contain the actual IP of the Splunk Machine and not Localhost
        self.splunk_host = splunk_variables.get("splunk_integration")["host"]

        # Combine the results in a valid URL that we can query splunk with
        self.splunk_instance = splunk_variables.get("splunk_integration")["url"]

    # Force the query to be in the Splunk StackPack with the yard id as the instance
    # This reduces boilerplate that has to be written on the queries
    def stackpack_topology_query(self, query_suffix: str):
        query_prefix = f"label IN ('stackpack:splunk', 'splunk-instance:{self.yard_id}')"
        return f"{query_prefix} AND {query_suffix}"


class SplunkCommonBase:
    def __init__(self, ansible_var, request, splunk_instance, splunk_user, splunk_pass):
        # The calculated splunk instance
        self.splunk_instance = splunk_instance

        # Create a session to control all requests
        self.session = requests.Session()
        self.session.verify = False

        # Authentication Details
        self.splunk_user = splunk_user
        self.splunk_pass = splunk_pass

        # Addons
        self.request = request
        self.ansible_var = ansible_var

        # Disable Security Warning
        requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

    @staticmethod
    def get_or_else(value: Optional[any] = None, alternative: [any] = None):
        if value is not None:
            return value
        else:
            return alternative

    def post_to_services_receivers_simple(self,
                                          json_data,
                                          param_data=None) -> str:
        data_dump_request_filename = "{}-{}_request.json".format(
            Path(str(self.request.node.fspath)).stem,
            self.request.node.originalname,
        )

        with open(data_dump_request_filename, "w") as outfile:
            json.dump({
                "json": json_data,
                "params": param_data
            }, outfile, indent=4)

        response = self.session.post("%s/services/receivers/simple" % self.splunk_instance,
                                     json=json_data,
                                     params=param_data,
                                     auth=(self.splunk_user, self.splunk_pass))

        data_dump_response_filename = "{}-{}_response.xml".format(
            Path(str(self.request.node.fspath)).stem,
            self.request.node.originalname,
        )

        xml: Element = ElementTree.fromstring(response.content)
        doc = ElementTree.SubElement(xml.find("results"), "doc")
        ElementTree.SubElement(doc, "http_response").text = "{}".format(response.status_code)
        root: ElementTree = ElementTree.ElementTree(xml)
        root.write(data_dump_response_filename)

        return response.raise_for_status()


class SplunkTopologyComponent(TypedDict):
    id: str
    type: str
    description: str
    topo_type: str


class SplunkTopologyRelation(TypedDict):
    relation_type: str
    source_id: str
    target_id: str
    description: str
    topo_type: str


class SplunkTopologyBase(SplunkCommonBase):

    def publish_component(self,
                          version: str = "v2",
                          component_id: Optional[str] = None,
                          component_type: Optional[str] = None,
                          component_description: Optional[str] = None,
                          component_topo_type: Optional[str] = None) -> SplunkTopologyComponent:
        # Create a type safe structure with the component we are psoting
        component = SplunkTopologyComponent(
            id=self.get_or_else(component_id, "server_{}".format(random.randint(0, 10000))),
            type=self.get_or_else(component_type, "server"),
            topo_type=self.get_or_else(component_topo_type, "component"),
            description=self.get_or_else(component_description, "Topology Server Component"),
        )

        self.post_to_services_receivers_simple(
            json_data={
                "version": version,
                "id": component.get("id"),
                "type": component.get("type"),
                "topo_type": component.get("topo_type"),
                "description": component.get("description")
            }
        )

        logging.debug(f"Publishing component with the name '{component.get('id')}' on StackState")

        return component

    def publish_relation(self,
                         source_id: str,
                         target_id: str,
                         version: str = "v2",
                         relation_type: Optional[str] = None,
                         description: Optional[str] = None,
                         topo_type: Optional[str] = None) -> SplunkTopologyRelation:
        # Create a type safe structure with the relation we are posting
        relation = SplunkTopologyRelation(
            source_id=source_id,
            target_id=target_id,
            relation_type=self.get_or_else(relation_type, "CONNECTED"),
            description=self.get_or_else(description, "Relation Description"),
            topo_type=self.get_or_else(topo_type, "relation"),
        )

        self.post_to_services_receivers_simple(
            json_data={
                "version": version,
                "topo_type": relation.get("topo_type"),
                "type": relation.get("relation_type"),
                "sourceId": relation.get("source_id"),
                "targetId": relation.get("target_id"),
                "description": relation.get("description")
            }
        )

        logging.debug(f"Publishing relation from '{relation.get('source_id')}' to '{relation.get('target_id')}'"
                      f" on StackState")

        return relation


class SplunkHealth(TypedDict):
    name: str
    check_state_id: str
    health: str
    topology_element_identifier: str
    message: Optional[str]


class SplunkHealthBase(SplunkCommonBase):
    # Core method for posting events to Splunk
    def publish_health(self,
                       version: str = "v2",
                       name: Optional[str] = None,
                       check_state_id: Optional[str] = None,
                       health: Optional[str] = None,
                       topology_element_identifier: Optional[str] = None,
                       message: Optional[str] = None) -> SplunkHealth:
        # Prepare the data that will be sent to StackState
        random_disk_id = random.randint(0, 10000)
        random_server_id = random.randint(0, 10000)

        # Create a type safe structure with the health we are posting
        health = SplunkHealth(
            name=self.get_or_else(name, "Disk {} SDA".format(random_disk_id)),
            check_state_id=self.get_or_else(check_state_id, "disk_{}_sda".format(random_disk_id)),
            health=self.get_or_else(health, random.choice(["CLEAR", "CRITICAL"])),
            message=self.get_or_else(message, "SDA Disk {} Message".format(random_disk_id)),
            topology_element_identifier=self.get_or_else(topology_element_identifier,
                                                         "server_{}".format(random_server_id)),
        )

        self.post_to_services_receivers_simple(
            json_data={
                "version": version,
                "check_state_id": health.get("check_state_id"),
                "name": health.get("name"),
                "health": health.get("health"),
                "topology_element_identifier": health.get("topology_element_identifier"),
                "message": health.get("message"),
            }
        )

        return health


class SplunkEvent(TypedDict):
    host: str
    source_type: str
    status: str
    description: str


class SplunkEventBase(SplunkCommonBase):
    # Core method for posting events to Splunk
    def publish_event(self,
                      version: str = "v2",
                      host: Optional[str] = None,
                      source_type: Optional[str] = None,
                      status: Optional[str] = None,
                      description: Optional[str] = None):
        # Prepare the data that will be sent to StackState
        random_host_id = random.randint(0, 10000)

        # Create a type safe structure with the health we are posting
        event = SplunkEvent(
            host=self.get_or_else(host, "host{}".format(random_host_id)),
            source_type=self.get_or_else(source_type, "sts_test_data"),
            status=self.get_or_else(status, random.choice(["CRITICAL", "OK"])),
            description=self.get_or_else(description, "Test host{} Event".format(random_host_id))
        )

        self.post_to_services_receivers_simple(
            json_data={
                "version": version,
                "status": event.get("status"),
                "description": event.get("description")
            },
            param_data={
                "host": event.get("host"),
                "sourcetype": event.get("source_type")
            }
        )

        return event


class SplunkMetric(TypedDict):
    host: str
    source_type: str
    topo_type: str
    metric: str
    value: str
    qa: str


class SplunkMetricBase(SplunkCommonBase):
    # Core method for posting events to Splunk
    def publish_metric(self,
                       version: str = "v2",
                       host: Optional[str] = None,
                       source_type: Optional[str] = None,
                       topo_type: Optional[str] = None,
                       metric: Optional[str] = None,
                       value: Optional[int] = None,
                       qa: Optional[str] = None) -> SplunkMetric:
        # Prepare the data that will be sent to StackState
        random_host_id = random.randint(0, 10000)

        # Create a type safe structure with the health we are posting
        metric = SplunkMetric(
            host=self.get_or_else(host, "host{}".format(random_host_id)),
            source_type=self.get_or_else(source_type, "sts_test_data"),
            topo_type=self.get_or_else(topo_type, "metrics"),
            metric=self.get_or_else(metric, "raw.metrics"),
            value=self.get_or_else(value, random.randint(1, 100000)),
            qa=self.get_or_else(qa, "splunk"),
        )

        self.post_to_services_receivers_simple(
            json_data={
                "version": version,
                "topo_type": metric.get("topo_type"),
                "metric": metric.get("metric"),
                "value": int(metric.get("value")),
                "qa": metric.get("qa")
            },
            param_data={
                "host": metric.get("host"),
                "sourcetype": metric.get("source_type")
            }
        )

        return metric
