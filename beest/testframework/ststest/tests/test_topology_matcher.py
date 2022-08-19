import random
from unittest import TestCase

from ststest import TopologyMatcher
from .fixtures import *


class TestTopologyMatcher(TestCase):

    def test_simple_positive(self):
        input_topology = TopologyFixture("a,b,c,a>before>b,c>before>b")
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(input_topology.topology())
        match = result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)

        self.assertEqual(input_topology.get('a'), match.component('A'))
        self.assertEqual(input_topology.get('b'), match.component('B'))

    def test_repeated(self):
        node1 = f"node-{random.randint(100, 200)}"
        node2 = f"node-{random.randint(100, 200)}"
        input_topology = TopologyFixture(f"cluster,{node1},{node2},{node1}>belongs>cluster,{node2}>belongs>cluster")
        matcher = TopologyMatcher() \
            .component("C", name=r"cluster") \
            .repeated(2,
                      lambda m: m
                      .component("N", name=r"node-\d+")
                      .one_way_direction("N", "C", type="belongs")
                      )

        result = matcher.find(input_topology.topology())
        match = result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)

        self.assertEqual(input_topology.get('cluster'), match.component('C'))
        self.assertUnorderedComponents(
            [input_topology.get(node1), input_topology.get(node2)],
            match.repeated_components('N')
        )

    def test_repeated_complex(self):
        node1 = f"node-{random.randint(100, 200)}"
        node2 = f"node-{random.randint(100, 200)}"
        daemonset = "my-daemonset"
        configmap = "my-cm"
        pod1 = f"{daemonset}-pod-{node1}"
        pod2 = f"{daemonset}-pod-{node2}"
        cluster = "cluster"

        input_topology = TopologyFixture(','.join([
            node1, node2, pod1, pod2,
            daemonset, configmap, cluster,
            f"{pod1}>scheduled_on>{node1}",
            f"{pod2}>scheduled_on>{node2}",
            f"{node1}>runs_on>{cluster}",
            f"{node2}>runs_on>{cluster}",
            f"{pod1}>uses>{configmap}",
            f"{pod2}>uses>{configmap}",
            f"{daemonset}>controls>{pod1}",
            f"{daemonset}>controls>{pod2}",
        ]))

        matcher = TopologyMatcher() \
            .component("CLUSTER", name=r"^cluster") \
            .component("DaemonSet", name=r"^my-daemonset$") \
            .component("ConfigMap", name=r"^my-cm$") \
            .repeated(2,
                      lambda m: m
                      .component("DsPod", name=r"my-daemonset-pod-.*")
                      .component("Node", name=r"^node-\d+")
                      .one_way_direction("DsPod", "Node", type="scheduled_on")
                      .one_way_direction("DsPod", "ConfigMap", type="uses")
                      .one_way_direction("DaemonSet", "DsPod", type="controls")
                      .one_way_direction("Node", "CLUSTER", type="runs_on")
                      )

        result = matcher.find(input_topology.topology())
        match = result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)

        self.assertEqual(input_topology.get('cluster'), match.component('CLUSTER'))
        self.assertEqual(input_topology.get('my-daemonset'), match.component('DaemonSet'))
        self.assertEqual(input_topology.get('my-cm'), match.component('ConfigMap'))
        self.assertUnorderedComponents(
            [input_topology.get(node1), input_topology.get(node2)],
            match.repeated_components('Node')
        )
        self.assertUnorderedComponents(
            [input_topology.get(pod1), input_topology.get(pod2)],
            match.repeated_components('DsPod')
        )

    def test_complex_positive(self):
        input_topology = TopologyFixture("a1,a2,b1,b2,c1,c2,a1>to>b1,a1>to>b2,a2>to>b1,b1>to>a2,b2>to>c1")
        matcher = TopologyMatcher() \
            .component("A", name="a.") \
            .component("B", name="b.") \
            .component("C", name="c.") \
            .one_way_direction("A", "B", type="to") \
            .one_way_direction("B", "C", type="to")

        result = matcher.find(input_topology.topology())
        match = result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)

        self.assertEqual(input_topology.get('a1'), match.component('A'))
        self.assertEqual(input_topology.get('b2'), match.component('B'))
        self.assertEqual(input_topology.get('c1'), match.component('C'))

    def test_ambiguous_match_failure(self):
        input_topology = TopologyFixture("a1,a2,b1,b2,c1,c2,a1>to>b2,b2>to>c1,a2>to>b2,b1>to>c1,b1>to>c2")
        matcher = TopologyMatcher() \
            .component("A", name="a.") \
            .component("B", name="b.") \
            .component("C", name="c.") \
            .one_way_direction("A", "B", type="to") \
            .one_way_direction("B", "C", type="to")

        result = matcher.find(input_topology.topology())
        with self.assertRaises(AssertionError) as exc:
            result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)

        exception_message = str(exc.exception)
        self.assertEqual(exception_message,
                         """
desired topology was not matched:
	multiple matches for component A[name~=a.]:
		#1#[a1](type=component,identifiers=)
		#2#[a2](type=component,identifiers=)
	multiple matches for component B[name~=b.]:
		#3#[b1](type=component,identifiers=)
		#4#[b2](type=component,identifiers=)
	multiple matches for component C[name~=c.]:
		#5#[c1](type=component,identifiers=)
		#6#[c2](type=component,identifiers=)
	multiple matches for relation A->B[type~=to,dependencyDirection~=ONE_WAY]:
		#2->[type=to]->4
		#1->[type=to]->4
	multiple matches for relation B->C[type~=to,dependencyDirection~=ONE_WAY]:
		#4->[type=to]->5
		#3->[type=to]->5
		#3->[type=to]->6
""".strip())

    def test_simple_wrong_relation(self):
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(TopologyFixture("a,b,c,b>before>c,a>after>b").topology())

        with self.assertRaises(AssertionError) as exc:
            result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)
        exception_message = str(exc.exception)

        self.assertEqual(exception_message,
                         """
desired topology was not matched:
	relation A->B[type~=before,dependencyDirection~=ONE_WAY] was not found
 """.strip())

    def test_topology_fixture(self):
        self.assertEqual(TopologyResult(
            components=[component_fixture(1, 'a')],
            relations=[]
        ), TopologyFixture("a").topology())

        self.assertEqual(TopologyResult(
            components=[component_fixture(1, 'a', outgoing=[3]),
                        component_fixture(2, 'b', incoming=[3])],
            relations=[relation_fixture(3, 1, 2, 'to')]
        ), TopologyFixture("a,b,a>to>b").topology())

        self.assertEqual(TopologyResult(
            components=[component_fixture(1, 'a', incoming=[6], outgoing=[3]),
                        component_fixture(2, 'b', incoming=[3, 7], outgoing=[5]),
                        component_fixture(4, 'c', incoming=[5], outgoing=[6, 7])],
            relations=[relation_fixture(3, 1, 2, 'to'),
                       relation_fixture(5, 2, 4, 'goes'),
                       relation_fixture(6, 4, 1, 'backto'),
                       relation_fixture(7, 4, 2, 'alsoto')]
        ), TopologyFixture("a,b,a>to>b,c,b>goes>c,c>backto>a,c>alsoto>b").topology())

        with self.assertRaises(KeyError, msg="forward reference should be rejected in constructor"):
            TopologyFixture("a,a>to>b,b")

    def assertUnorderedComponents(self, expected: list[ComponentWrapper], actual: list[ComponentWrapper]):
        self.assertEqual(
            sorted(expected, key=lambda c: c.id),
            sorted(actual, key=lambda c: c.id)
        )
