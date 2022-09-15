import pydot
import hashlib
import logging

from typing import Tuple

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper

from .matches.topology_match import TopologyMatch
from .primitive_matchers import ComponentMatcher, RelationMatcher
from .match_keys import ComponentKey, RelationKey


class ExactMatches:
    matcher_to_relation: dict[RelationKey, any]
    relation_to_matcher: dict[any, RelationKey]
    matcher_to_component: dict[ComponentKey, any]
    component_to_matcher: dict[any, ComponentKey]

    def __init__(self,
                 matcher_to_relation: dict[RelationKey, any],
                 relation_to_matcher: dict[any, RelationKey],
                 matcher_to_component: dict[ComponentKey, any],
                 component_to_matcher: dict[any, ComponentKey],
                 ):
        self.component_to_matcher = component_to_matcher
        self.matcher_to_component = matcher_to_component
        self.relation_to_matcher = relation_to_matcher
        self.matcher_to_relation = matcher_to_relation


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

    @staticmethod
    def _component_matcher_label(matcher: 'ComponentMatcher'):
        rules = "\n".join([str(m) for m in matcher.matchers])
        return f"{matcher.id}\n{rules}"

    @staticmethod
    def _relation_matcher_label(matcher: 'RelationMatcher'):
        return "\n".join([str(m) for m in matcher.matchers])

    @staticmethod
    def _compute_exact_matches(topology_matches: list[TopologyMatch],
                               component_matches: dict[str, list[ComponentWrapper]],
                               relation_matches: dict[str, list[RelationWrapper]]):
        if len(topology_matches) == 1:
            topo_match = topology_matches[0]
            return ExactMatches(
                matcher_to_component={mkey: comp.id for mkey, comp in topo_match.components.items()},
                component_to_matcher={comp.id: mkey for mkey, comp in topo_match.components.items()},
                matcher_to_relation={mkey: rel.id for mkey, rel in topo_match.relations.items()},
                relation_to_matcher={rel.id: mkey for mkey, rel in topo_match.relations.items()},
            )
        else:
            matcher_to_component = {mkey: comps[0].id for mkey, comps in component_matches.items() if len(comps) == 1}
            matcher_to_relation = {mkey: rels[0].id for mkey, rels in relation_matches.items() if len(rels) == 1}
            return ExactMatches(
                matcher_to_component=matcher_to_component,
                matcher_to_relation=matcher_to_relation,
                component_to_matcher={c: m for m, c in matcher_to_component.items()},
                relation_to_matcher={r: m for m, r in matcher_to_relation.items()},
            )

    def render_debug_dot(self,
                         topology_matches: list[TopologyMatch],
                         source: TopologyResult,
                         component_matchers: list[ComponentMatcher],
                         component_matchers_index: dict[ComponentKey, ComponentMatcher],
                         component_matches: dict[str, list[ComponentWrapper]],
                         relation_matchers: list[RelationMatcher],
                         relation_matchers_index: dict[Tuple[ComponentKey, ComponentKey], RelationMatcher],
                         relation_matches: dict[str, list[RelationWrapper]],
                         matching_graph_name: str = None
                         ):
        graph = pydot.Dot("Topology match debug", graph_type="digraph")

        # we need to know which component/relations are matched exactly to overlay them on the diagram
        exact_matches = self._compute_exact_matches(topology_matches=topology_matches,
                                                    component_matches=component_matches,
                                                    relation_matches=relation_matches)

        # graph_name should cluster_{i}, otherwise the renderer does not recognize styles
        query_graph = pydot.Subgraph(graph_name="cluster_1", label="Query result", **self.QueryResultSubgraphStyle)
        for component in source.components:
            component_gv_id = component.id
            label = f"{component.name}\ntype={component.type}"
            color = 'black'
            if component.id in exact_matches.component_to_matcher:
                color = DotGraphDrawer.ExactMatchColor
                matcher = component_matchers_index[exact_matches.component_to_matcher[component.id]]
                label += f"\n-----\n{self._component_matcher_label(matcher)}"
            query_graph.add_node(pydot.Node(component_gv_id, label=label, color=color))
        for relation in source.relations:
            relation_gv_id = relation.id
            color = 'black'
            label = relation.type
            if relation.id in exact_matches.relation_to_matcher:
                color = DotGraphDrawer.ExactMatchColor
                matcher = relation_matchers_index[exact_matches.relation_to_matcher[relation.id]]
                label += f"\n-----\n{self._relation_matcher_label(matcher)}"
            self._add_compound_relation(
                query_graph, relation_gv_id, relation.source, relation.target, color,
                shape="underline", label=label)
        graph.add_subgraph(query_graph)

        matcher_graph = pydot.Subgraph(graph_name="cluster_0", label="Matching rule",
                                       **self.MatchingRuleSubgraphStyle)
        for matcher in component_matchers:
            if matcher.id in exact_matches.matcher_to_component:
                continue
            matcher_gv_id = f"{matcher.id}_matcher"
            matches = component_matches.get(matcher.id, [])
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

        for matcher in relation_matchers:
            if matcher.id in exact_matches.matcher_to_relation:
                continue
            matcher_gv_id = str(matcher.id)
            label = self._relation_matcher_label(matcher)
            matches = relation_matches.get(matcher.id, [])
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
