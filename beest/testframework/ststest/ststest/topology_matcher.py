import hashlib
import logging
from typing import Optional

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper, TopologyDeleteWrapper, \
    TopologyStartSnapshotWrapper, TopologyStopSnapshotWrapper
import pydot
import urllib.parse
import pyshorteners

from .primitive_matchers import ComponentMatcher, RelationMatcher, DeleteMatcher, StartSnapshotMatcher, \
    StopSnapshotMatcher
from .invariant_search import ConsistentGraphMatcher


class TopologyMatch:
    def __init__(self, components: dict[str, ComponentWrapper], relations: dict[str, RelationWrapper],
                 deletes: dict[str, TopologyDeleteWrapper], start_snapshot: Optional[TopologyStartSnapshotWrapper],
                 stop_snapshot: Optional[TopologyStopSnapshotWrapper]):
        self._components = components
        self._relations = relations
        self._deletes = deletes
        self._start_snapshot = start_snapshot
        self._stop_snapshot = stop_snapshot

    def __repr__(self):
        components = "\n\t".join([f"{key}: {comp}" for key, comp in self._components.items()])
        relations = "\n\t".join([f"{source} > {target}: {comp}" for (source, target), comp in self._relations.items()])
        deletes = "\n\t".join([f"{key}: {comp}" for key, comp in self._deletes.items()])
        start_snapshot = "\n\t" + str(self._start_snapshot) + "\n\t" if self._start_snapshot else ''
        stop_snapshot = "\n\t" + str(self._stop_snapshot) + "\n\t" if self._stop_snapshot else ''

        return "Match[\n\t" \
               + start_snapshot \
               + components \
               + "\n\t" \
               + relations \
               + "\n\t" \
               + deletes \
               + stop_snapshot \
               + "\n]"

    def __eq__(self, other):
        if isinstance(other, TopologyMatch):
            return self._components == other._components and \
                   self._relations == other._relations and \
                   self._deletes == other._deletes and \
                   self._start_snapshot == other._start_snapshot and \
                   self._stop_snapshot == other._stop_snapshot
        return False

    def component(self, key: str) -> ComponentWrapper:
        return self._components.get(key)

    def has_component(self, id: int) -> bool:
        return next((True for comp in self._components.values() if comp.id == id), False)

    def has_relation(self, id: int) -> bool:
        return next((True for rel in self._relations.values() if rel.id == id), False)

    def delete(self, key) -> TopologyDeleteWrapper:
        return self._deletes.get(key)

    def start_snapshot(self) -> Optional[TopologyStartSnapshotWrapper]:
        return self._start_snapshot

    def stop_snapshot(self) -> Optional[TopologyStopSnapshotWrapper]:
        return self._stop_snapshot


class TopologyMatchingResult:
    def __init__(self,
                 matches: list[TopologyMatch],
                 matcher: 'TopologyMatcher',
                 source: TopologyResult,
                 component_matches: dict[str, list[ComponentWrapper]],
                 relation_matches: dict[str, list[RelationWrapper]],
                 delete_matches: dict[str, list[TopologyDeleteWrapper]],
                 start_snapshot_match: Optional[TopologyStartSnapshotWrapper],
                 stop_snapshot_match: Optional[TopologyStopSnapshotWrapper]
                 ):
        self._topology_matches = matches
        self._relation_matches = relation_matches
        self._component_matches = component_matches
        self._delete_matches = delete_matches
        self._start_snapshot_match = start_snapshot_match
        self._stop_snapshot_match = stop_snapshot_match
        self._matcher = matcher
        self._source = source

    @staticmethod
    def component_pretty_short(comp: ComponentWrapper):
        # TODO print attributes related to a matcher
        return f"#{comp.id}#[{comp.name}](type={comp.type},identifiers={','.join(map(str, comp.attributes.get('identifiers', [])))})"

    @staticmethod
    def relation_pretty_short(rel: RelationWrapper):
        # TODO print attributes related to a matcher
        return f"#{rel.source}->[type={rel.type}]->{rel.target}"

    def _assert_single_match(self, matches, matcher_dict) -> list[str]:
        errors = []
        delimiter = "\n\t\t"

        for key, items in matches.items():
            matcher = matcher_dict[key]
            if len(items) == 0:
                errors.append(f"\t{matcher.matcher_type()} {matcher} was not found")
            elif len(items) > 1:
                errors.append(f"\tmultiple matches for {matcher.matcher_type()} {matcher}:"
                              f"{delimiter}{delimiter.join(map(self.component_pretty_short, items))}")

        return errors

    def assert_exact_match(self, matching_graph_name=None, matching_graph_upload=True) -> TopologyMatch:
        if len(self._topology_matches) == 1:
            return self._topology_matches[0]
        errors = []

        # component matchers
        comp_matchers = {matcher.id: matcher for matcher in self._matcher.components}
        errors = errors + self._assert_single_match(self._component_matches, comp_matchers)

        # relation matchers
        rel_matchers = {matcher.id(): matcher for matcher in self._matcher.relations}
        errors = errors + self._assert_single_match(self._relation_matches, rel_matchers)

        # delete matchers
        del_matchers = {matcher.id(): matcher for matcher in self._matcher.deletes}
        errors = errors + self._assert_single_match(self._delete_matches, del_matchers)

        # start snapshot match
        if self._start_snapshot_match:
            if not self._matcher._start_snapshot:
                errors.append(f"\t{self._matcher._start_snapshot.matcher_type()} "
                              f"{self._matcher._start_snapshot} was not found")

        # stop snapshot match
        if self._stop_snapshot_match:
            if not self._matcher._stop_snapshot:
                errors.append(f"\t{self._matcher._stop_snapshot.matcher_type()} "
                              f"{self._matcher._stop_snapshot} was not found")

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
        for mcomp in self._source.components:
            label = f"{mcomp.name}"
            color = 'black'
            if exact_match is not None and exact_match.has_component(mcomp.id):
                color = 'darkgreen'
            query_graph.add_node(pydot.Node(mcomp.id, label=label, color=color))
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
        for mcomp in self._matcher.components:
            id = mcomp.id
            rules = "\n".join([str(m) for m in mcomp.matchers])
            label = f"{id}\n{rules}"
            matches = self._component_matches.get(id, [])
            color = self._color_for_matches_count(len(matches))
            matcher_graph.add_node(pydot.Node(id, label=label, color=color))
            for comp in matches:
                graph.add_edge(pydot.Edge(id, comp.id, color=color, style="dotted", penwidth=5))

        for rel in self._matcher.relations:
            rules = "\n".join([str(m) for m in rel.matchers])

            matches = self._relation_matches.get(rel.id(), [])
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


def get_common_relations(sources: list[ComponentWrapper], targets: list[ComponentWrapper]):
    # TODO consider BOTH_WAY type of relations
    source_relations = set([id for source in sources for id in source.outgoing_relations])
    target_relations = set([id for target in targets for id in target.incoming_relations])
    return list(source_relations & target_relations)


class TopologyMatcher:
    def __init__(self):
        self.components: list[ComponentMatcher] = []
        self.relations: list[RelationMatcher] = []
        self.deletes: list[DeleteMatcher] = []
        self.start_snapshot: Optional[StartSnapshotMatcher] = None
        self.stop_snapshot: Optional[StopSnapshotMatcher] = None

    def component(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.components.append(ComponentMatcher(id, kwargs))
        return self

    def one_way_direction(self, source: str, target: str, **kwargs) -> 'TopologyMatcher':
        source_found = False
        target_found = False
        for comp in self.components:
            if comp.id == source:
                source_found = True
            if comp.id == target:
                target_found = True
        if not source_found:
            raise KeyError(f"source `{source}` have not been found, use .component('{source}') to define a component "
                           f"before defining a relation")
        if not target_found:
            raise KeyError(f"target `{target}` have not been found, use .component('{target}') to define a component "
                           f"before defining a relation")

        kwargs['dependencyDirection'] = 'ONE_WAY'
        self.relations.append(RelationMatcher(source, target, kwargs))
        return self

    def find(self, topology: TopologyResult) -> TopologyMatchingResult:
        component_by_id = {comp.id: comp for comp in topology.components}
        relation_by_id = {rel.id: rel for rel in topology.relations}

        errors = []

        def add_error(message):
            errors.append(message)

        consistent_graph_matcher = ConsistentGraphMatcher()

        # find all matching components and group them by virtual node (id) from a pattern
        matching_components: dict[str, list[ComponentWrapper]] = {}
        for comp_match in self.components:
            matching_components[comp_match.id] = [comp for comp in topology.components if comp_match.match(comp)]

        # tell CGM that for every virtual node (A) there is a list of possible options (A1..An)
        for key, component_candidates in matching_components.items():
            consistent_graph_matcher.add_choice_of_spec([{key: comp.id} for comp in component_candidates])

        # now we are looking for relations (e.g. A1>B2..Ax>By) that possibly represents a defined relation A>B
        matching_relations = {}
        for comp_rel in self.relations:
            source_candidates = matching_components.get(comp_rel.source, [])
            target_candidates = matching_components.get(comp_rel.target, [])
            relation_candidate_ids = get_common_relations(source_candidates, target_candidates)
            relation_candidates = [relation_by_id[id] for id in relation_candidate_ids if id in relation_by_id]
            matching = [rel for rel in relation_candidates if comp_rel.match(rel)]
            matching_relations[comp_rel.id()] = matching
            consistent_graph_matcher.add_choice_of_spec([
                {
                    comp_rel.source: rel.source,
                    comp_rel.target: rel.target,
                    comp_rel.id(): rel.id,
                }
                for rel in matching
            ])

        def build_topo_match_from_spec(spec: dict) -> TopologyMatch:
            components = {}
            relations = {}
            for key, id in spec.items():
                if id in relation_by_id:
                    relations[key] = relation_by_id[id]
                else:
                    components[key] = component_by_id[id]
            return TopologyMatch(components, relations)

        result_graph_specs = consistent_graph_matcher.get_graphs()
        return TopologyMatchingResult(
            list(map(build_topo_match_from_spec, result_graph_specs)),
            self,
            topology,
            matching_components,
            matching_relations,
        )
