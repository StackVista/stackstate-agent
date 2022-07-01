import util
from ststest import TopologyMatcher

testinfra_hosts = ["local"]


def test_lambda_topology_is_present(cliv1):
    lambda_api_topology = \
        TopologyMatcher() \
        .component("gateway_api", name=r"^.*-rest-api$") \
        .component("gateway_stage", name=r"^(.*)-rest-api - (.*)-test$") \
        .component("gateway_resource", name=r"^/\{proxy\+\}$") \
        .component("gateway_method", name=r"^GET$") \
        .component("lambda", name=r".*-hello$") \
        .one_way_direction("gateway_api", "gateway_stage", type="has resource") \
        .one_way_direction("gateway_stage", "gateway_resource", type="uses service") \
        .one_way_direction("gateway_resource", "gateway_method", type="uses service") \
        .one_way_direction("gateway_method", "lambda", type="uses service")

    def assert_it():
        topology = cliv1.topology(
            "layer in ('Serverless') AND environment in ('Production')")
        match_result = lambda_api_topology.find(topology)
        match_result.assert_exact_match()

    util.wait_until(assert_it, 5, 1)
