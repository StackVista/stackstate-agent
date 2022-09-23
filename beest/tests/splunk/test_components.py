import util

from typing import Optional
from splunk_testing_base import SplunkBase, SplunkTopologyComponent
from agent_tesing_base import AgentTestingBase
from conftest import YARD_LOCATION
from stscliv1 import CLIv1
from ststest import TopologyMatcher

# Create a connection through a specific inventory host
# When running the script outside Beest we need a relative location for ansible_inventory file.
# This works inside the Beest container and outside Beest so this can be as is for both.
testinfra_hosts = [f"ansible://local?ansible_inventory={YARD_LOCATION}/ansible_inventory"]


def test_splunk_component(agent: AgentTestingBase,
                          splunk: SplunkBase,
                          cliv1: CLIv1,
                          simulator):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    component: SplunkTopologyComponent = splunk.topology.publish_component()

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher()\
            .component(component.get("id"), name=component.get("id"), type=component.get("type"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{component.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


def test_splunk_multiple_component(agent: AgentTestingBase,
                                   splunk: SplunkBase,
                                   cliv1: CLIv1,
                                   simulator):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    # Make sure the agent is running
    agent.start_agent_on_host()

    # Publish a Splunk Component to the Splunk Instance to be used in testing
    component_a: SplunkTopologyComponent = splunk.topology.publish_component()
    component_b: SplunkTopologyComponent = splunk.topology.publish_component()
    component_c: SplunkTopologyComponent = splunk.topology.publish_component()

    # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
    def topology_matcher():
        return TopologyMatcher() \
            .component(component_a.get("id"), name=component_a.get("id"), type=component_a.get("type")) \
            .component(component_b.get("id"), name=component_b.get("id"), type=component_b.get("type")) \
            .component(component_c.get("id"), name=component_c.get("id"), type=component_c.get("type"))

    # Wait until we find this component in StackState. If it does not succeed after x seconds then we will dump the
    # simulator logs if it is available.
    util.wait_until_topology_match(
        cliv1,
        topology_matcher=topology_matcher,
        topology_query=lambda: f"name = '{component_a.get('id')}' OR "
                               f"name = '{component_b.get('id')}' OR "
                               f"name = '{component_c.get('id')}'",
        timeout=120,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
        period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
        on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
    )


# Stateful State
# We will publish a component while the agent is active and wait for it
# When we find it then we will stop the agent and post a second component
# After a few minutes we start the agent up again
# And wait to find the second component
def test_splunk_component_stateful_state(agent: AgentTestingBase,
                                         cliv1: CLIv1,
                                         simulator,
                                         splunk: SplunkBase):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_component(expect_failure: bool = False,
                       expected_component: SplunkTopologyComponent = None) -> Optional[SplunkTopologyComponent]:
        component: SplunkTopologyComponent = expected_component

        if component is None:
            component = splunk.topology.publish_component()

        try:
            # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
            def topology_matcher():
                return TopologyMatcher().component(component.get("id"),
                                                   name=component.get("id"),
                                                   type=component.get("type"))

            # Wait until we find this component in StackState. If it does not succeed after x
            # seconds then we will dump the simulator logs if it is available.
            util.wait_until_topology_match(
                cliv1,
                topology_matcher=topology_matcher,
                topology_query=lambda: f"name = '{component.get('id')}'",
                timeout=80,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
                period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
                on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
            )
        except Exception as e:
            if expect_failure is True:
                return component
            else:
                raise e

        if expect_failure is True:
            raise Exception("Component should not exist but did not fail with a exception")
        else:
            return component

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    component_posted_while_agent_was_down: Optional[SplunkTopologyComponent] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def post_while_agent_is_stopped():
        nonlocal component_posted_while_agent_was_down
        component_posted_while_agent_was_down = post_component(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def post_after_agent_started():
        post_component(expected_component=component_posted_while_agent_was_down)
        post_component()

    # Run a stateful test for the agent
    agent.stateful_state_run_cycle_test(
        func_before_agent_stop=post_component,
        func_after_agent_stop=post_while_agent_is_stopped,
        func_after_agent_startup=post_after_agent_started
    )


# Transactional State
# We will produce a component while the routing is open
# Then we will close the routes, post another component and make sure that the component does not exist
# After that we will open the routes and test if the component eventually end up in STS
def test_splunk_component_transactional_check(agent: AgentTestingBase,
                                              cliv1: CLIv1,
                                              splunk: SplunkBase,
                                              simulator):
    # Make sure we have routing enabled
    agent.allow_routing_to_sts_instance()

    def post_component(expect_failure: bool = False,
                       expected_component: SplunkTopologyComponent = None) -> Optional[SplunkTopologyComponent]:
        component: SplunkTopologyComponent = expected_component

        if component is None:
            component = splunk.topology.publish_component()

        try:
            # The topology_matcher process that will be executed every x seconds in the wait_until_topology_match cycle
            def topology_matcher():
                return TopologyMatcher().component(component.get("id"),
                                                   name=component.get("id"),
                                                   type=component.get("type"))

            # Wait until we find this component in StackState. If it does not succeed after x
            # seconds then we will dump the simulator logs if it is available.
            util.wait_until_topology_match(
                cliv1,
                topology_matcher=topology_matcher,
                topology_query=lambda: f"name = '{component.get('id')}'",
                timeout=80,  # Run for a total of x seconds, Sometimes the Agent check can take some time so to be safe
                period=5,  # Run the 'topology_matcher' and 'topology_query' every x seconds
                on_failure_action=lambda: simulator()  # Dump the simulator logs if the cycle failed (If enabled)
            )
        except Exception as e:
            if expect_failure is True:
                return component
            else:
                raise e

        if expect_failure is True:
            raise Exception("Component should not exist but did not fail with a exception")
        else:
            return component

    # A component that was posted while the agent was stopped, this should not exist after it starts up again
    component_after_blocking_routes: Optional[SplunkTopologyComponent] = None

    # Post a component while the agent is stopped, when then assign this to a variable to test again after wards
    def post_while_routes_is_blocked():
        nonlocal component_after_blocking_routes
        component_after_blocking_routes = post_component(expect_failure=True)

    # Attempt to check the prev component we posted should be in the agent including the
    # new one we posted
    def post_while_routes_is_open():
        post_component(expected_component=component_after_blocking_routes)
        post_component()

    # Run a stateful test for the agent
    agent.transactional_run_cycle_test(
        func_before_blocking_routes=lambda: post_component(),
        func_after_blocking_routes=lambda: post_while_routes_is_blocked(),
        rerun_func_unblocking_blocked_routes=lambda: post_while_routes_is_open()
    )
