from unittest import TestCase
from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper
from ststest import TopologyMatcher, TopologyMatch


def component_fixture(id: int, name: str, outgoing: list[int] = None, incoming=None) -> ComponentWrapper:
    return ComponentWrapper({
        'id': id, 'name': name, 'type': 'component',
        'incomingRelations': incoming if incoming is not None else [],
        'outgoingRelations': outgoing if outgoing is not None else [],
    })


def relation_fixture(id: int, source: int, target: int, type: str, direction: str = 'ONE_WAY') -> RelationWrapper:
    return RelationWrapper({
        'id': id,
        'source': source,
        'target': target,
        'type': type,
        'dependencyDirection': direction,
    })


class TopologyFixture:
    def __init__(self, definition: str):
        self.elements = {}
        self.components = []
        self.relations = []

        element_id = 1
        component_ids = {}
        for element_definition in definition.split(","):
            elem_def_parts = element_definition.split(">")
            if len(elem_def_parts) == 1:
                name = elem_def_parts[0]
                component = component_fixture(element_id, name)
                component_ids[name] = component
                self.elements[element_definition] = component
                self.components.append(component)
            elif len(elem_def_parts) == 3:
                source = component_ids[elem_def_parts[0]]
                type = elem_def_parts[1]
                target = component_ids[elem_def_parts[2]]
                relation = relation_fixture(element_id, source.id, target.id, type)
                self.elements[element_definition] = relation
                self.relations.append(relation)
                source.attributes['outgoingRelations'].append(element_id)
                target.attributes['incomingRelations'].append(element_id)
            else:
                raise ValueError(f"unrecognized definition `{element_definition}`, should either `some_name` or `some_a>relates>some_b`")
            element_id += 1

    def get(self, definition):
        return self.elements.get(definition)

    def topology(self):
        return TopologyResult(self.components, self.relations)


class TestTopologyMatcher(TestCase):
    def test_topology_fixture(self):
        self.assertEqual(TopologyResult(
            components=[
                component_fixture(1, 'a')
            ],
            relations=[]
            ), TopologyFixture("a").topology())

        self.assertEqual(TopologyResult(
            components=[
                component_fixture(1, 'a', outgoing=[3]),
                component_fixture(2, 'b', incoming=[3]),
            ],
            relations=[
                relation_fixture(3, 1, 2, 'to'),
            ]
        ), TopologyFixture("a,b,a>to>b").topology())

        self.assertEqual(TopologyResult(
            components=[
                component_fixture(1, 'a', incoming=[6], outgoing=[3]),
                component_fixture(2, 'b', incoming=[3], outgoing=[5]),
                component_fixture(4, 'c', incoming=[5], outgoing=[6]),
            ],
            relations=[
                relation_fixture(3, 1, 2, 'to'),
                relation_fixture(5, 2, 4, 'goes'),
                relation_fixture(6, 4, 1, 'backto'),
            ]
        ), TopologyFixture("a,b,a>to>b,c,b>goes>c,c>backto>a").topology())

    def test_simple_positive(self):
        input_topology = TopologyFixture("a,b,c,a>before>b")
        matcher = TopologyMatcher() \
            .component("A", name="a") \
            .component("B", name="b") \
            .one_way_direction("A", "B", type="before")

        result = matcher.find(input_topology.topology())

        self.assertEqual(0, len(result.errors), msg="no errors are expected")
        self.assertEqual([TopologyMatch(
            components={
                'A': input_topology.get('a'),
                'B': input_topology.get('b'),
            },
            relations={}  # TODO relations are not computed yet
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
                components={
                    'A': input_topology.get('a1'),
                    'B': input_topology.get('b2'),
                    'C': input_topology.get('c1'),
                },
                relations={}  # TODO relations are not computed yet
            ),
            TopologyMatch(
                components={
                    'A': input_topology.get('a2'),
                    'B': input_topology.get('b2'),
                    'C': input_topology.get('c1'),
                },
                relations={}  # TODO relations are not computed yet
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
        # self.assertEqual([], result.matches, msg="should not have found anything")
