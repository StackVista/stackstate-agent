import pydot
import hashlib
import logging
import urllib.parse
import pyshorteners

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper

from .topology_match import TopologyMatch
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
    ExactMatchColor = 'darkgreen'

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

        exact_match = None
        if len(topology_matches) == 1:
            exact_match = topology_matches[0]

        # graph_name should cluster_{i}, otherwise the renderer does not recognize styles
        query_graph = pydot.Subgraph(graph_name="cluster_1", label="Query result", **self.QueryResultSubgraphStyle)
        for mcomp in source.components:
            label = f"{mcomp.name}"
            color = 'black'
            if exact_match is not None and exact_match.has_component(mcomp.id):
                color = 'darkgreen'
            query_graph.add_node(pydot.Node(mcomp.id, label=label, color=color))
        for rel in source.relations:
            relation_node_id = rel.id
            color = 'black'
            if exact_match is not None and exact_match.has_relation(rel.id):
                color = 'darkgreen'
            self._add_compound_relation(
                query_graph, relation_node_id, rel.source, rel.target, color,
                shape="underline", label=rel.type)
        graph.add_subgraph(query_graph)

        matcher_graph = pydot.Subgraph(graph_name="cluster_0", label="Matching rule", **self.MatchingRuleSubgraphStyle)
        for mcomp in component_matchers:
            id = mcomp.id
            rules = "\n".join([str(m) for m in mcomp.matchers])
            label = f"{id}\n{rules}"
            matches = component_matches.get(id, [])
            color = self._color_for_matches_count(len(matches))
            matcher_graph.add_node(pydot.Node(id, label=label, color=color))
            for comp in matches:
                graph.add_edge(pydot.Edge(id, comp.id, color=color, style="dotted", penwidth=5))

        for rel in relation_matchers:
            rules = "\n".join([str(m) for m in rel.matchers])

            matches = relation_matches.get(rel.id(), [])
            color = self._color_for_matches_count(len(matches))
            self._add_compound_relation(
                matcher_graph, rel.id(), rel.source, rel.target, color,
                shape="underline", label=rules)
            # connect to matched relations
            for mrel in matches:
                graph.add_edge(pydot.Edge(rel.id(), mrel.id, color=color, style="dotted", penwidth=3))

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
            except Exception:
                logging.warning("could not make matching diagram available at URL", exc_info=True)
