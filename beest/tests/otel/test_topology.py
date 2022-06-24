import util
from topology_assertion import TopologyMatcher

testinfra_hosts = ["local"]


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


def test_lambda_topology_is_present(cliv1):
    def assert_it():
        topology = cliv1.topology(
            "layer in ('Serverless') AND environment in ('Production')")
        lambda_topology = lambda_api_topology.find(topology)

    util.wait_until(assert_it, 5, 1)


def test_gitlab_runner(cliv1):
    topology = cliv1.topology(
        "layer in ('Serverless') AND environment in ('Production')")
    gitlab_runner_topology = TopologyMatcher().component('grl', name="windows-gitlab-runner7-function", type="Lambda Function")
    gitlab_runner = gitlab_runner_topology.find(topology).components['grl']
    runner_telemetry = cliv1.telemetry([gitlab_runner['id']])
    assert False, str(runner_telemetry)
