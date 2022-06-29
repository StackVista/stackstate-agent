from unittest import TestCase
from ststest import TopologyMatcher, TopologyMatch
from fixtures import *


class TestTopologyMatcher(TestCase):

    def test_simple_positive(self):
        input_topology = TopologyFixture("a,b,c,a>before>b")
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(input_topology.topology())

        self.assertEqual(0, len(result.errors), msg="no errors are expected")
        self.assertEqual([TopologyMatch(
            components={'A': input_topology.get('a'),
                        'B': input_topology.get('b')},
            relations={('A', 'B'): input_topology.get('a>before>b')}
        )], result.matches)

    def test_finds_within_multiple_matches(self):
        input_topology = TopologyFixture("a1,a2,b1,b2,c1,c2,a1>to>b2,b2>to>c1,a2>to>b2,b1>to>c1,b1>to>c2")
        matcher = TopologyMatcher() \
            .component("A", name="a.") \
            .component("B", name="b.") \
            .component("C", name="c.") \
            .one_way_direction("A", "B", type="to") \
            .one_way_direction("B", "C", type="to")

        result = matcher.find(input_topology.topology())

        self.assertEqual([], result.errors, msg="no errors are expected")
        self.assertEqual([
            TopologyMatch(
                components={'A': input_topology.get('a1'),
                            'B': input_topology.get('b2'),
                            'C': input_topology.get('c1')},
                relations={('A', 'B'): input_topology.get('a1>to>b2'),
                           ('B', 'C'): input_topology.get('b2>to>c1')}
            ),
            TopologyMatch(
                components={'A': input_topology.get('a2'),
                            'B': input_topology.get('b2'),
                            'C': input_topology.get('c1')},
                relations={('A', 'B'): input_topology.get('a2>to>b2'),
                           ('B', 'C'): input_topology.get('b2>to>c1')}
            ),
        ], result.matches)

    def test_simple_wrong_direction(self):
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(TopologyFixture("a,b,c,b>before>c").topology())

        self.assertEqual(1, len(result.errors), msg="no errors expected")
        self.assertRegex(result.errors[0], r"relation.*A.*B")

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
