import pydot
import hashlib
import logging

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper

from ststest.ststest.matches.topology_match import TopologyMatch
from .primitive_matchers import ComponentMatcher, RelationMatcher


class DotGraphDrawer:
    QueryResultSubgraphStyle = {
        'fontsize': 30,
        'color': 'mediumslateblue',
        'penwidth': 5,
    }
    MatchingRuleSubgraphStyle = {
        'fontsize': 30,
        'color': 'grey',
        'penwidth': 5,
    }
    UnmatchedColor = 'red'
    MultipleMatches = 'orange'
    ExactMatchColor = 'green'

    def __init__(self):
        pass

    @staticmethod
    def _color_for_matches_count(count: int):
        if count == 1:
            return DotGraphDrawer.ExactMatchColor
        elif count == 0:
            return DotGraphDrawer.UnmatchedColor
        else:
            return DotGraphDrawer.MultipleMatches

    @staticmethod
    def _add_compound_relation(graph: pydot.Subgraph, id, source, target, color, **kwargs):
        graph.add_node(pydot.Node(id, **kwargs, color=color))
        graph.add_edge(pydot.Edge(source, id, color=color))
        graph.add_edge(pydot.Edge(id, target, color=color))

    def _component_matcher_label(self, matcher: 'ComponentMatcher'):
        rules = "\n".join([str(m) for m in matcher.matchers])
        return f"{matcher.id}\n{rules}"

    def _relation_matcher_label(self, matcher: 'RelationMatcher'):
        return "\n".join([str(m) for m in matcher.matchers])

    def render_debug_dot(self,
                         topology_matches: list[TopologyMatch],
                         source: TopologyResult,
                         component_matchers: list[ComponentMatcher],
                         component_matches: dict[str, list[ComponentWrapper]],
                         relation_matchers: list[RelationMatcher],
                         relation_matches: dict[str, list[RelationWrapper]],
                         matching_graph_name: str = None,
                         generate_diagram_url: bool = True
                         ):
        graph = pydot.Dot("Topology match debug", graph_type="digraph")

        # we need to know which component/relations are matched exactly to overlay them on the diagram
        exact_matches = self._compute_exact_matches()

        # graph_name should cluster_{i}, otherwise the renderer does not recognize styles
        query_graph = pydot.Subgraph(graph_name="cluster_1", label="Query result", **self.QueryResultSubgraphStyle)
        for component in self._source.components:
            component_gv_id = component.id
            label = f"{component.name}\ntype={component.type}"
            color = 'black'
            if component.id in exact_matches.component_to_matcher:
                color = DotGraphDrawer.ExactMatchColor
                matcher = self._component_matchers_index[exact_matches.component_to_matcher[component.id]]
                label += f"\n-----\n{self._component_matcher_label(matcher)}"
            query_graph.add_node(pydot.Node(component_gv_id, label=label, color=color))
        for relation in self._source.relations:
            relation_gv_id = relation.id
            color = 'black'
            label = relation.type
            if relation.id in exact_matches.relation_to_matcher:
                color = DotGraphDrawer.ExactMatchColor
                matcher = self._relation_matchers_index[exact_matches.relation_to_matcher[relation.id]]
                label += f"\n-----\n{self._relation_matcher_label(matcher)}"
            self._add_compound_relation(
                query_graph, relation_gv_id, relation.source, relation.target, color,
                shape="underline", label=label)
        graph.add_subgraph(query_graph)

        matcher_graph = pydot.Subgraph(graph_name="cluster_0", label="Matching rule",
                                       **self.MatchingRuleSubgraphStyle)
        for matcher in self._matcher._components:
            if matcher.id in exact_matches.matcher_to_component:
                continue
            matcher_gv_id = f"{matcher.id}_matcher"
            matches = self._component_matches.get(matcher.id, [])
            color = self._color_for_matches_count(len(matches))
            matcher_graph.add_node(
                pydot.Node(matcher_gv_id, label=self._component_matcher_label(matcher), color=color))
            for comp in matches:
                graph.add_edge(pydot.Edge(matcher_gv_id, comp.id, color=color, style="dotted", penwidth=5))

        def component_gv_id(comp_matcher: ComponentKey):
            if comp_matcher in exact_matches.matcher_to_component:
                # connect relation to the exact component that was matched by the matcher
                return exact_matches.matcher_to_component[comp_matcher]
            return f"{comp_matcher}_matcher"

        for matcher in self._matcher._relations:
            if matcher.id in exact_matches.matcher_to_relation:
                continue
            matcher_gv_id = str(matcher.id)
            label = self._relation_matcher_label(matcher)
            matches = self._relation_matches.get(matcher.id, [])
            color = self._color_for_matches_count(len(matches))
            self._add_compound_relation(
                matcher_graph, matcher_gv_id, component_gv_id(matcher.source), component_gv_id(matcher.target),
                color,
                shape="underline", label=label)
            # connect to matched relations
            for mrel in matches:
                graph.add_edge(pydot.Edge(matcher_gv_id, str(mrel.id), color=color, style="dotted", penwidth=3))

        graph.add_subgraph(matcher_graph)

        graph_dot_str = graph.to_string()

        if matching_graph_name is not None:
            dot_file = matching_graph_name + ".gv"
        else:
            dot_file = hashlib.sha1(graph_dot_str.encode('utf-8')).hexdigest()[0:10] + ".gv"
        with open(dot_file, 'w') as dfp:
            dfp.write(graph_dot_str)
            logging.info("saved match in a DOT file at %s", dot_file)
