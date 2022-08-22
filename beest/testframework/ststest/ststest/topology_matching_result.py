import hashlib
import logging

from pyshorteners.exceptions import ShorteningErrorException

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper
import pydot
import urllib.parse
import pyshorteners

from .topology_match import TopologyMatch


class TopologyMatchingResult:

    def __init__(self,
                 matches: list[TopologyMatch],
                 matcher: 'TopologyMatcher',
                 source: TopologyResult,
                 component_matches: dict[str, list[ComponentWrapper]],
                 relation_matches: dict[str, list[RelationWrapper]],
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
        rel_matchers = {matcher.id(): matcher for matcher in self._matcher._relations}
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

    def render_debug_dot(self, matching_graph_name=None, generate_diagram_url=True):
        graph = pydot.Dot("Topology match debug", graph_type="digraph")

        exact_match = None
        if len(self._topology_matches) == 1:
            exact_match = self._topology_matches[0]

        # graph_name should cluster_{i}, otherwise the renderer does not recognize styles
        query_graph = pydot.Subgraph(graph_name="cluster_1", label="Query result", **self.QueryResultSubgraphStyle)
        for scomp in self._source.components:
            label = f"{scomp.name}\ntype={scomp.type}"
            color = 'black'
            if exact_match is not None and exact_match.has_component(scomp.id):
                color = 'darkgreen'
            query_graph.add_node(pydot.Node(scomp.id, label=label, color=color))
        for rel in self._source.relations:
            relation_node_id = rel.id
            color = 'black'
            if exact_match is not None and exact_match.has_relation(rel.id):
                color = 'darkgreen'
            self._add_compound_relation(
                query_graph, relation_node_id, rel.source, rel.target, color,
                shape="underline", label=rel.type)
        graph.add_subgraph(query_graph)

        matcher_graph = pydot.Subgraph(graph_name="cluster_0", label="Matching rule", **self.MatchingRuleSubgraphStyle)
        for mcomp in self._matcher._components:
            id = str(mcomp.id)
            rules = "\n".join([str(m) for m in mcomp.matchers])
            label = f"{id}\n{rules}"
            matches = self._component_matches.get(id, [])
            color = self._color_for_matches_count(len(matches))
            matcher_graph.add_node(pydot.Node(id, label=label, color=color))
            for comp in matches:
                graph.add_edge(pydot.Edge(id, comp.id, color=color, style="dotted", penwidth=5))

        for rel in self._matcher._relations:
            rel_id = str(rel.id())
            rules = "\n".join([str(m) for m in rel.matchers])
            matches = self._relation_matches.get(rel.id(), [])
            color = self._color_for_matches_count(len(matches))
            self._add_compound_relation(
                matcher_graph, rel_id, str(rel.source), str(rel.target), color,
                shape="underline", label=rules)
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
