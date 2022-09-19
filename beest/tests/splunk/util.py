import time

from typing import Callable
from stscliv1 import CLIv1
from ststest import TopologyMatcher


def wait_until_topology_match(cliv1: CLIv1,
                              topology_matcher: Callable[[], TopologyMatcher],
                              topology_query: Callable[[], str],
                              timeout: int,
                              period: int) -> None:
    def loop():
        # Call the Lambda function to get the Topology builder and the compiled query
        expected = topology_matcher()
        actual = topology_query()

        # Run the CLI and the TopologyBuilder and attempt to find a match
        actual_topology = cliv1.topology(actual)
        expected_matches = expected.find(actual_topology)
        expected_matches.assert_exact_match()

    wait_until(loop, timeout, period)


def wait_until(someaction: Callable[[any, any], any],
               timeout: int,
               period: int = 0.25,
               *args: any,
               **kwargs: any):
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
