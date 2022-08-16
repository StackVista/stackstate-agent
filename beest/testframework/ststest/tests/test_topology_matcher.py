import random
from itertools import permutations
from typing import Callable, TypeVar
from unittest import TestCase
from ststest import TopologyMatcher, TopologyMatch
from .fixtures import *

T = TypeVar('T')


def assert_permuted(elements: list[any], code: Callable[[any], any]):
    last_error = None
    for input in permutations(elements):
        try:
            code(*input)
            return
        except AssertionError as e:
            last_error = e
    raise last_error


class TestTopologyMatcher(TestCase):

    def test_simple_positive(self):
        input_topology = TopologyFixture("a,b,c,a>before>b,c>before>b")
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(input_topology.topology())
        match = result.assert_exact_match(matching_graph_name=self._testMethodName, matching_graph_upload=False)
        self.assertEqual(TopologyMatch(
            components={'A': input_topology.get('a'),
                        'B': input_topology.get('b')},
            relations={('A', 'B'): input_topology.get('a>before>b')}
        ), match)

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

        def assert_option(node1, node2):
            self.assertEqual(TopologyMatch(
                components={'C': input_topology.get('cluster'),
                            'N.0': input_topology.get(node1),
                            'N.1': input_topology.get(node2)},
                relations={('N.0', 'C'): input_topology.get(f'{node1}>belongs>cluster'),
                           ('N.1', 'C'): input_topology.get(f'{node2}>belongs>cluster')
                           }
            ), match)
            self.assertEqual(input_topology.get('cluster'), match.component('C'))
            assert_permuted([node1, node2],
                            lambda node1, node2:
                            self.assertEqual(
                                [input_topology.get(node1), input_topology.get(node2)],
                                match.components('N'))
                            )

        assert_permuted([node1, node2], assert_option)

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
        self.assertEqual(TopologyMatch(
            components={'A': input_topology.get('a1'),
                        'B': input_topology.get('b2'),
                        'C': input_topology.get('c1')},
            relations={('A', 'B'): input_topology.get('a1>to>b2'),
                       ('B', 'C'): input_topology.get('b2>to>c1')}
        ), match)

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
