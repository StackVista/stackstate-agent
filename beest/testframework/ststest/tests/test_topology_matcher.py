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
        match = result.assert_exact_match(matching_graph_name=self._testMethodName)

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
        match = result.assert_exact_match(matching_graph_name=self._testMethodName)

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
        match = result.assert_exact_match(matching_graph_name=self._testMethodName)

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
        match = result.assert_exact_match(matching_graph_name=self._testMethodName)

        self.assertEqual(input_topology.get('a1'), match.component('A'))
        self.assertEqual(input_topology.get('b2'), match.component('B'))
        self.assertEqual(input_topology.get('c1'), match.component('C'))

    def test_render_dot(self):
        input_topology = TopologyFixture(
            'a,b,c1,c2,d1,d2,e,a>to>b,b>to>c1,b>to>c2,c1>to>d1,c2>to>d2,d1>to>e,d2>to>e,f1,f2,e>to>f1,e>to>f2'
        )
        matcher = TopologyMatcher() \
            .component('A', name='a') \
            .component('B', name='b') \
            .one_way_direction('A', 'B') \
            .component('C1', name='c1') \
            .component('C2', name='c_miss') \
            .component('D1', name='d1') \
            .component('D2', name='d2') \
            .one_way_direction('B', 'C2') \
            .one_way_direction('C2', 'D2') \
            .one_way_direction('C1', 'D1', type='cd_miss') \
            .component('E', name='e') \
            .component('F', name='f.') \
            .one_way_direction('E', 'F')

        expected_dot_file = '''digraph "Topology match debug" {
subgraph cluster_1 {
color=mediumslateblue;
fontsize=30;
label="Query result";
penwidth=5;
1 [color=green, label="a\\ntype=component\\n-----\\nA\\nname~=a"];
2 [color=green, label="b\\ntype=component\\n-----\\nB\\nname~=b"];
3 [color=green, label="c1\\ntype=component\\n-----\\nC1\\nname~=c1"];
4 [color=black, label="c2\\ntype=component"];
5 [color=green, label="d1\\ntype=component\\n-----\\nD1\\nname~=d1"];
6 [color=green, label="d2\\ntype=component\\n-----\\nD2\\nname~=d2"];
7 [color=green, label="e\\ntype=component\\n-----\\nE\\nname~=e"];
15 [color=black, label="f1\\ntype=component"];
16 [color=black, label="f2\\ntype=component"];
8 [color=green, label="to\\n-----\\ndependencyDirection~=ONE_WAY", shape=underline];
1 -> 8  [color=green];
8 -> 2  [color=green];
9 [color=black, label=to, shape=underline];
2 -> 9  [color=black];
9 -> 3  [color=black];
10 [color=black, label=to, shape=underline];
2 -> 10  [color=black];
10 -> 4  [color=black];
11 [color=black, label=to, shape=underline];
3 -> 11  [color=black];
11 -> 5  [color=black];
12 [color=black, label=to, shape=underline];
4 -> 12  [color=black];
12 -> 6  [color=black];
13 [color=black, label=to, shape=underline];
5 -> 13  [color=black];
13 -> 7  [color=black];
14 [color=black, label=to, shape=underline];
6 -> 14  [color=black];
14 -> 7  [color=black];
17 [color=black, label=to, shape=underline];
7 -> 17  [color=black];
17 -> 15  [color=black];
18 [color=black, label=to, shape=underline];
7 -> 18  [color=black];
18 -> 16  [color=black];
}

F_matcher -> 15  [color=orange, penwidth=5, style=dotted];
F_matcher -> 16  [color=orange, penwidth=5, style=dotted];
"('E', 'F')" -> 17  [color=orange, penwidth=3, style=dotted];
"('E', 'F')" -> 18  [color=orange, penwidth=3, style=dotted];
subgraph cluster_0 {
color=grey;
fontsize=30;
label="Matching rule";
penwidth=5;
C2_matcher [color=red, label="C2\\nname~=c_miss"];
F_matcher [color=orange, label="F\\nname~=f."];
"('B', 'C2')" [color=red, label="dependencyDirection~=ONE_WAY", shape=underline];
2 -> "('B', 'C2')"  [color=red];
"('B', 'C2')" -> C2_matcher  [color=red];
"('C2', 'D2')" [color=red, label="dependencyDirection~=ONE_WAY", shape=underline];
C2_matcher -> "('C2', 'D2')"  [color=red];
"('C2', 'D2')" -> 6  [color=red];
"('C1', 'D1')" [color=red, label="type~=cd_miss\\ndependencyDirection~=ONE_WAY", shape=underline];
3 -> "('C1', 'D1')"  [color=red];
"('C1', 'D1')" -> 5  [color=red];
"('E', 'F')" [color=orange, label="dependencyDirection~=ONE_WAY", shape=underline];
7 -> "('E', 'F')"  [color=orange];
"('E', 'F')" -> F_matcher  [color=orange];
}

}
'''

        result = matcher.find(input_topology.topology())
        with self.assertRaises(AssertionError) as exc:
            result.assert_exact_match(matching_graph_name=self._testMethodName)
        with open(f"{self._testMethodName}.gv", 'r') as dot:
            test = ''.join(dot.readlines())
            self.assertEqual(expected_dot_file, test)

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
            result.assert_exact_match(matching_graph_name=self._testMethodName)

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
            result.assert_exact_match(matching_graph_name=self._testMethodName)
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
