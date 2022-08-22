import hashlib
import logging

from pyshorteners.exceptions import ShorteningErrorException

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper
import pydot
import urllib.parse
import pyshorteners

from .match_keys import ComponentKey
from .topology_match import TopologyMatch, RelationKey


class TopologyMatchingResult:

    _matcher: 'TopologyMatcher'

    def __init__(self,
                 matches: list[TopologyMatch],
                 matcher: 'TopologyMatcher',
                 source: TopologyResult,
                 component_matches: dict[ComponentKey, list[ComponentWrapper]],
                 relation_matches: dict[RelationKey, list[RelationWrapper]],
                 ):
        self._topology_matches = matches
        self._relation_matches = relation_matches
        self._component_matches = component_matches
        self._matcher = matcher
        self._source = source

    @staticmethod
    def component_pretty_short(comp: ComponentWrapper):
        # TODO print attributes related to a matcher
        properties = f"type={comp.type},identifiers={','.join(map(str, comp.attributes.get('identifiers', [])))}"
        return f"#{comp.id}#[{comp.name}]({properties})"

    @staticmethod
    def relation_pretty_short(rel: RelationWrapper):
        # TODO print attributes related to a matcher
        return f"#{rel.source}->[type={rel.type}]->{rel.target}"

    def assert_exact_match(self, matching_graph_name=None, matching_graph_upload=True) -> TopologyMatch:
        if len(self._topology_matches) == 1:
            return self._topology_matches[0]
        errors = []
        delimiter = "\n\t\t"
        comp_matchers = {matcher.id: matcher for matcher in self._matcher._components}
        for key, components in self._component_matches.items():
            matcher = comp_matchers[key]
            if len(components) == 0:
                errors.append(f"\tcomponent {matcher} was not found")
            elif len(components) > 1:
                errors.append(f"\tmultiple matches for component {matcher}:"
                              f"{delimiter}{delimiter.join(map(self.component_pretty_short, components))}")
        rel_matchers = {matcher.id: matcher for matcher in self._matcher._relations}
        for key, relations in self._relation_matches.items():
            matcher = rel_matchers[key]
            if len(relations) == 0:
                errors.append(f"\trelation {matcher} was not found")
            elif len(relations) > 1:
                errors.append(f"\tmultiple matches for relation {matcher}:"
                              f"{delimiter}{delimiter.join(map(self.relation_pretty_short, relations))}")
        self.render_debug_dot(matching_graph_name, matching_graph_upload)
        error_sep = "\n"
        assert False, f"desired topology was not matched:\n{error_sep.join(errors)}"

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
    ExactMatchColor = 'darkgreen'

    @staticmethod
    def _color_for_matches_count(count: int):
        if count == 1:
            return TopologyMatchingResult.ExactMatchColor
        elif count == 0:
            return TopologyMatchingResult.UnmatchedColor
        else:
            return TopologyMatchingResult.MultipleMatches

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

    def _get_comp_matcher_by_key(self, key):
        for m in self._matcher._components:
            if m.id == key:
                return m
        return None

    def _get_rel_matcher_by_key(self, key):
        for m in self._matcher._relations:
            if m.id == key:
                return m
        return None

    def render_debug_dot(self, matching_graph_name=None, generate_diagram_url=True):
        graph = pydot.Dot("Topology match debug", graph_type="digraph")

        exact_match = None
        if len(self._topology_matches) == 1:
            exact_match = self._topology_matches[0]

        # inverted index of matchers for specific component
        exact_component_matches = {}
        for mkey, comps in self._component_matches.items():
            for comp in comps:
                matchers = exact_component_matches[comp.id] if comp.id in exact_component_matches else []
                matchers.append(mkey)
                exact_component_matches[comp.id] = matchers

        exact_relation_matches = {}
        for mkey, rels in self._relation_matches.items():
            for rel in rels:
                matchers = exact_relation_matches[rel.id] if rel.id in exact_component_matches else []
                matchers.append(mkey)
                exact_relation_matches[rel.id] = matchers

        exact_matchers = set()
        for _, matchers in exact_component_matches.items():
            if len(matchers) == 1:
                exact_matchers.add(matchers[0])


        # graph_name should cluster_{i}, otherwise the renderer does not recognize styles
        query_graph = pydot.Subgraph(graph_name="cluster_1", label="Query result", **self.QueryResultSubgraphStyle)
        for scomp in self._source.components:
            label = f"{scomp.name}\ntype={scomp.type}"
            color = 'black'
            if scomp.id in exact_component_matches and len(exact_component_matches[scomp.id]) == 1:
                color = 'darkgreen'
                matcher = self._get_comp_matcher_by_key(exact_component_matches[scomp.id][0])
                label += f"\n-----\n{self._component_matcher_label(matcher)}"
            query_graph.add_node(pydot.Node(scomp.id, label=label, color=color))
        for rel in self._source.relations:
            relation_node_id = rel.id
            color = 'black'
            label = rel.type
            if rel.id in exact_relation_matches and len(exact_relation_matches[rel.id]) == 1:
                color = 'darkgreen'
                matcher = self._get_rel_matcher_by_key(exact_relation_matches[rel.id][0])
                label += f"\n-----\n{self._relation_matcher_label(matcher)}"
            self._add_compound_relation(
                query_graph, relation_node_id, rel.source, rel.target, color,
                shape="underline", label=label)
        graph.add_subgraph(query_graph)

        matcher_graph = pydot.Subgraph(graph_name="cluster_0", label="Matching rule", **self.MatchingRuleSubgraphStyle)
        for mcomp in self._matcher._components:
            id = str(mcomp.id)
            matches = self._component_matches.get(id, [])
            color = self._color_for_matches_count(len(matches))
            if len(matches) != 1:
                matcher_graph.add_node(pydot.Node(id, label=self._component_matcher_label(mcomp), color=color))
                for comp in matches:
                    graph.add_edge(pydot.Edge(id, comp.id, color=color, style="dotted", penwidth=5))

        for rel in self._matcher._relations:
            rel_id = str(rel.id)
            label = self._relation_matcher_label(rel)
            matches = self._relation_matches.get(rel.id, [])
            color = self._color_for_matches_count(len(matches))
            if len(matches) == 1:
                continue
            source = str(rel.source) if rel.source not in exact_matchers else self._component_matches[rel.source][0].id
            target = str(rel.target) if rel.target not in exact_matchers else self._component_matches[rel.target][0].id
            self._add_compound_relation(
                matcher_graph, rel_id, source, target, color,
                shape="underline", label=label)
            # connect to matched relations
            for mrel in matches:
                graph.add_edge(pydot.Edge(rel_id, str(mrel.id), color=color, style="dotted", penwidth=3))

        graph.add_subgraph(matcher_graph)

        graph_dot_str = graph.to_string()

        if matching_graph_name is not None:
            dot_file = matching_graph_name + ".gv"
        else:
            dot_file = hashlib.sha1(graph_dot_str.encode('utf-8')).hexdigest()[0:10] + ".gv"
        with open(dot_file, 'w') as dfp:
            dfp.write(graph_dot_str)
            logging.info("saved match in a DOT file at %s", dot_file)

        if not generate_diagram_url:
            logging.info("matching diagram was not request (generate_diagram_url=False)")
        else:
            try:
                base_share_url = urllib.parse.urlparse("https://graphviz.sandbox.stackstate.io/")
                share_url = base_share_url._replace(fragment=urllib.parse.quote(graph_dot_str))
                shortened_url = pyshorteners.Shortener().tinyurl.short(share_url.geturl())
                logging.info("matching diagram is available at %s", shortened_url)
            except ShorteningErrorException:
                logging.warning("could not make matching diagram available at URL", exc_info=True)
