from typing import Optional

import util

from agent_tesing_base import AgentTestingBase
from splunk_testing_base import SplunkBase, SplunkTopologyComponent, SplunkTopologyRelation
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_relations(splunk: SplunkBase,
                          cliv1: CLIv1,
                          agent: AgentTestingBase,
                          simulator):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    source: SplunkTopologyComponent = splunk.topology.publish_component()
    target: SplunkTopologyComponent = splunk.topology.publish_component()

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target.get("id"))

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def relation_matcher():
        return TopologyMatcher()\
            .component(source.get("id"), name=source.get("id"), type="server")\
            .component(target.get("id"), name=target.get("id"), type="server")\
            .one_way_direction(source=source.get("id"), target=target.get("id"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=relation_matcher,
        topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


def test_splunk_multiple_relation(splunk: SplunkBase,
                                  agent: AgentTestingBase,
                                  cliv1: CLIv1,
                                  simulator):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    source: SplunkTopologyComponent = splunk.topology.publish_component()
    target_a: SplunkTopologyComponent = splunk.topology.publish_component()
    target_b: SplunkTopologyComponent = splunk.topology.publish_component()

    # Publish a Splunk Relation to the Splunk Instance to be used in testing
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target_a.get("id"))
    splunk.topology.publish_relation(source_id=source.get("id"),
                                     target_id=target_b.get("id"))

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def relation_matcher():
        return TopologyMatcher()\
            .component(source.get("id"), name=source.get("id"), type="server")\
            .component(target_a.get("id"), name=target_a.get("id"), type="server")\
            .component(target_b.get("id"), name=target_b.get("id"), type="server")\
            .one_way_direction(source=source.get("id"), target=target_a.get("id"))\
            .one_way_direction(source=source.get("id"), target=target_b.get("id"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=relation_matcher,
        topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target_a.get('id')}' "
                               f"OR name = '{target_b.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


# Stateful State
# We will publish a relation while the agent is active and wait for it
# When we find it then we will stop the agent and post a second relation
# After a few minutes we start the agent up again
# And wait to find the second relation
def test_splunk_relation_stateful_state(agent: AgentTestingBase,
                                        cliv1: CLIv1,
                                        splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def find_relation_in_sts(expect_failure: bool = False,
                             expected_relation: SplunkTopologyRelation = None,
                             expected_source: SplunkTopologyComponent = None,
                             expected_target: SplunkTopologyComponent = None) -> [Optional[SplunkTopologyRelation],
                                                                                  Optional[SplunkTopologyComponent],
                                                                                  Optional[SplunkTopologyComponent]]:
        # Set internal state for manipulation
        relation: SplunkTopologyRelation = expected_relation
        source: SplunkTopologyComponent = expected_source
        target: SplunkTopologyComponent = expected_target

        # Create a new relation if any of the above is undefined
        # If the one above is defined then we are just attempting to retest if the relation exists
        if relation is None or source is None or target is None:
            # Publish the source component
            source: SplunkTopologyComponent = splunk.topology.publish_component()
            # Publish the target component
            target: SplunkTopologyComponent = splunk.topology.publish_component()
            # Now publish the relation between these components
            relation = splunk.topology.publish_relation(source_id=source.get("id"), target_id=target.get("id"))

        try:
            # The relation_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
            def relation_matcher():
                return TopologyMatcher() \
                    .component(source.get("id"), name=source.get("id"), type="server") \
                    .component(target.get("id"), name=target.get("id"), type="server") \
                    .one_way_direction(source=source.get("id"), target=target.get("id"))

            # Wait until we find this component in StackState. If it does not succeed after x
            # seconds then we will dump the simulator logs if it is available.
            util.wait_until_topology_match(
                cliv1,
                topology_matcher=relation_matcher,
                topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target.get('id')}'",
                timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
                period=10,  # Run the 'topology_matcher' and 'topology_query' every x seconds
            )
        except Exception as e:
            if expect_failure is True:
                return [source, target, relation]
            else:
                raise e

        if expect_failure is True:
            raise Exception("Relation should not exist but did not fail with a exception")
        else:
            return [source, target, relation]

    # A relation that was posted while the agent was stopped, this should not exist after it starts up again
    relation_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None
    source_component_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None
    target_component_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None

    # Post a relation while the agent is stopped, when then assign this to a variable to test again after wards
    def find_relation_while_agent_is_stopped():
        nonlocal relation_posted_while_agent_was_down
        nonlocal source_component_posted_while_agent_was_down
        nonlocal target_component_posted_while_agent_was_down

        relation_result = find_relation_in_sts(expect_failure=True)

        source_component_posted_while_agent_was_down = relation_result[0]
        target_component_posted_while_agent_was_down = relation_result[1]
        relation_posted_while_agent_was_down = relation_result[2]

    # Attempt to check the prev relation we posted should be in the agent including the
    # new one we posted
    def find_relation_after_agent_started():
        find_relation_in_sts(expected_relation=relation_posted_while_agent_was_down,
                             expected_source=source_component_posted_while_agent_was_down,
                             expected_target=target_component_posted_while_agent_was_down)

    # Run a stateful test for the agent
    agent.stateful_state_run_cycle_test(
        func_before_agent_stop=find_relation_in_sts,
        func_after_agent_stop=find_relation_while_agent_is_stopped,
        func_after_agent_startup=find_relation_after_agent_started
    )


# Transactional State
# We will produce a relation while the routing is open
# Then we will close the routes, post another relation and make sure that the relation does not exist
# After that we will open the routes and test if the relation eventually end up in STS
def test_splunk_relation_transactional_check(agent: AgentTestingBase,
                                             cliv1: CLIv1,
                                             splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def find_relation_in_sts(expect_failure: bool = False,
                             expected_relation: SplunkTopologyRelation = None,
                             expected_source: SplunkTopologyComponent = None,
                             expected_target: SplunkTopologyComponent = None) -> [Optional[SplunkTopologyRelation],
                                                                                  Optional[SplunkTopologyComponent],
                                                                                  Optional[SplunkTopologyComponent]]:
        # Set internal state for manipulation
        relation: SplunkTopologyRelation = expected_relation
        source: SplunkTopologyComponent = expected_source
        target: SplunkTopologyComponent = expected_target

        # Create a new relation if any of the above is undefined
        if relation is None or source is None or target is None:
            source: SplunkTopologyComponent = splunk.topology.publish_component()
            target: SplunkTopologyComponent = splunk.topology.publish_component()
            relation = splunk.topology.publish_relation(source_id=source.get("id"), target_id=target.get("id"))

        try:
            # The relation_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
            def relation_matcher():
                return TopologyMatcher() \
                    .component(source.get("id"), name=source.get("id"), type="server") \
                    .component(target.get("id"), name=target.get("id"), type="server") \
                    .one_way_direction(source=source.get("id"), target=target.get("id"))

            # Wait until we find this component in StackState. If it does not succeed after x
            # seconds then we will dump the simulator logs if it is available.
            util.wait_until_topology_match(
                cliv1,
                topology_matcher=relation_matcher,
                topology_query=lambda: f"name = '{source.get('id')}' OR name = '{target.get('id')}'",
                timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
                period=10,  # Run the 'topology_matcher' and 'topology_query' every x seconds
            )
        except Exception as e:
            if expect_failure is True:
                return [source, target, relation]
            else:
                raise e

        if expect_failure is True:
            raise Exception("Relation should not exist but did not fail with a exception")
        else:
            return [source, target, relation]

    # A relation that was posted while the agent was stopped, this should not exist after it starts up again
    relation_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None
    source_component_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None
    target_component_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def find_relation_while_routes_is_blocked():
        nonlocal relation_posted_while_agent_was_down
        nonlocal source_component_posted_while_agent_was_down
        nonlocal target_component_posted_while_agent_was_down

        relation_result = find_relation_in_sts(expect_failure=True)

        source_component_posted_while_agent_was_down = relation_result[0]
        target_component_posted_while_agent_was_down = relation_result[1]
        relation_posted_while_agent_was_down = relation_result[2]

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def find_relation_while_routes_is_open():
        find_relation_in_sts(expected_relation=relation_posted_while_agent_was_down,
                             expected_source=source_component_posted_while_agent_was_down,
                             expected_target=target_component_posted_while_agent_was_down)

    # Run a stateful test for the agent
    agent.transactional_run_cycle_test(
        func_before_blocking_routes=find_relation_in_sts,
        func_after_blocking_routes=find_relation_while_routes_is_blocked,
        rerun_func_unblocking_blocked_routes=find_relation_while_routes_is_open
    )
