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
        components = "\n".join([f"{key}: {comp}" for key, comp in self._components.items()])
        relations = "\n".join([f"{rel.source} > {rel.target}: {rel}" for _, rel in self._relations.items()])
        deletes = "\n".join([f"{key}: {dlt}" for key, dlt in self._deletes.items()])
        start_snapshot = "\n\t" + str(self._start_snapshot) + "\n\t" if self._start_snapshot else ''
        stop_snapshot = "\n\t" + str(self._stop_snapshot) + "\n\t" if self._stop_snapshot else ''

        return f"Match" \
               f"{start_snapshot}" \
               f"\n[Components]\n" \
               f"{components}" \
               f"\n[Relations]\n" \
               f"{relations}" \
               f"\n[Deletes]\n" \
               f"{deletes}" \
               f"{stop_snapshot}" \
               "\n"

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

    @staticmethod
    def _assert_single_match(matches, matcher_dict, printer) -> list[str]:
        errors = []
        delimiter = "\n\t\t"

        for key, items in matches.items():
            matcher = matcher_dict[key]
            if len(items) == 0:
                errors.append(f"\t{matcher.matcher_type()} {matcher} was not found")
            elif len(items) > 1:
                errors.append(f"\tmultiple matches for {matcher.matcher_type()} {matcher}:"
                              f"{delimiter}{delimiter.join(map(printer, items))}")

        return errors

    def assert_exact_match(self, matching_graph_name=None, matching_graph_upload=True) -> TopologyMatch:
        if len(self._topology_matches) == 1:
            return self._topology_matches[0]
        errors = []

        # component matchers
        comp_matchers = {matcher.id: matcher for matcher in self._matcher.component_matchers}
        errors = errors + self._assert_single_match(self._component_matches, comp_matchers, self.component_pretty_short)

        # relation matchers
        rel_matchers = {matcher.id(): matcher for matcher in self._matcher.relation_matchers}
        errors = errors + self._assert_single_match(self._relation_matches, rel_matchers, self.relation_pretty_short)

        # delete matchers
        del_matchers = {matcher.id: matcher for matcher in self._matcher.delete_matchers}
        errors = errors + self._assert_single_match(self._delete_matches, del_matchers, str)

        # start snapshot match
        if self._start_snapshot_match:
            if not self._matcher.start_snapshot_matcher:
                errors.append(f"\t{self._matcher.start_snapshot_matcher.matcher_type()} "
                              f"{self._matcher.start_snapshot_matcher} was not found")

        # stop snapshot match
        if self._stop_snapshot_match:
            if not self._matcher.stop_snapshot_matcher:
                errors.append(f"\t{self._matcher.stop_snapshot_matcher.matcher_type()} "
                              f"{self._matcher.stop_snapshot_matcher} was not found")

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
        for mcomp in self._matcher.component_matchers:
            id = mcomp.id
            rules = "\n".join([str(m) for m in mcomp.matchers])
            label = f"{id}\n{rules}"
            matches = self._component_matches.get(id, [])
            color = self._color_for_matches_count(len(matches))
            matcher_graph.add_node(pydot.Node(id, label=label, color=color))
            for comp in matches:
                graph.add_edge(pydot.Edge(id, comp.id, color=color, style="dotted", penwidth=5))

        for rel in self._matcher.relation_matchers:
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
        self.component_matchers: list[ComponentMatcher] = []
        self.relation_matchers: list[RelationMatcher] = []
        self.delete_matchers: list[DeleteMatcher] = []
        self.start_snapshot_matcher: Optional[StartSnapshotMatcher] = None
        self.stop_snapshot_matcher: Optional[StopSnapshotMatcher] = None

    def component(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.component_matchers.append(ComponentMatcher(id, kwargs))
        return self

    def start_snapshot(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.start_snapshot_matcher = StartSnapshotMatcher(id)
        return self

    def stop_snapshot(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.stop_snapshot_matcher = StopSnapshotMatcher(id)
        return self

    def delete(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.delete_matchers.append(DeleteMatcher(id, kwargs))
        return self

    def one_way_direction(self, source: str, target: str, **kwargs) -> 'TopologyMatcher':
        source_found = False
        target_found = False
        for comp in self.component_matchers:
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
        self.relation_matchers.append(RelationMatcher(source, target, kwargs))
        return self

    def _match_components(self, topology: TopologyResult,
                          cgm: ConsistentGraphMatcher) -> dict[str, list[ComponentWrapper]]:

        # find all matching components and group them by virtual node (id) from a pattern
        matching_components: dict[str, list[ComponentWrapper]] = {}
        for comp_match in self.component_matchers:
            matching_components[comp_match.id] = [comp for comp in topology.components if comp_match.match(comp)]

        # tell CGM that for every virtual node (A) there is a list of possible options (A1..An)
        for key, component_candidates in matching_components.items():
            cgm.add_choice_of_spec([{key: comp.id} for comp in component_candidates])

        return matching_components

    def _match_relations(self, relation_by_id: dict[str, RelationWrapper],
                         matching_components: dict[str, list[ComponentWrapper]],
                         cgm: ConsistentGraphMatcher) -> dict[str, list[RelationWrapper]]:
        # now we are looking for relations (e.g. A1>B2..Ax>By) that possibly represents a defined relation A>B
        matching_relations: dict[str, list[RelationWrapper]] = {}
        for comp_rel in self.relation_matchers:
            source_candidates = matching_components.get(comp_rel.source, [])
            target_candidates = matching_components.get(comp_rel.target, [])
            relation_candidate_ids = get_common_relations(source_candidates, target_candidates)
            relation_candidates = [relation_by_id[id] for id in relation_candidate_ids if id in relation_by_id]
            matching = [rel for rel in relation_candidates if comp_rel.match(rel)]
            matching_relations[comp_rel.id()] = matching
            cgm.add_choice_of_spec([
                {
                    comp_rel.source: rel.source,
                    comp_rel.target: rel.target,
                    comp_rel.id(): rel.id,
                }
                for rel in matching
            ])

        return matching_relations

    def _match_deletes(self, topology: TopologyResult,
                       cgm: ConsistentGraphMatcher) -> dict[str, list[TopologyDeleteWrapper]]:

        # find all matching deletes and group them by virtual node (id) from a pattern
        matching_deletes: dict[str, list[TopologyDeleteWrapper]] = {}
        for delete_match in self.delete_matchers:
            matching_deletes[delete_match.id] = [dlt for dlt in topology.deletes if delete_match.match(dlt)]

        # tell CGM that for every virtual node (A) there is a list of possible options (A1..An)
        for key, delete_candidates in matching_deletes.items():
            cgm.add_choice_of_spec([{key: dlt.id} for dlt in delete_candidates])

        return matching_deletes

    @staticmethod
    def _build_topo_match_from_cgm_spec(cgm_spec: dict,
                                        component_by_id: dict[str, ComponentWrapper],
                                        relation_by_id: dict[str, RelationWrapper],
                                        delete_by_id: dict[str, TopologyDeleteWrapper]) -> TopologyMatch:
        components = {}
        relations = {}
        deletes = {}
        for key, spec_id in cgm_spec.items():
            if spec_id in relation_by_id:
                relations[key] = relation_by_id[spec_id]
            elif spec_id in delete_by_id:
                deletes[key] = delete_by_id[spec_id]
            else:
                components[key] = component_by_id[spec_id]
        tm = TopologyMatch(components, relations, deletes, None, None)
        return tm

    def _match_graphs(self,
                      cgm: ConsistentGraphMatcher,
                      component_by_id: dict[str, ComponentWrapper],
                      relation_by_id: dict[str, RelationWrapper],
                      delete_by_id: dict[str, TopologyDeleteWrapper]) -> list[TopologyMatch]:

        result_graph_specs = cgm.get_graphs()

        matches: list[TopologyMatch] = []
        for spec in result_graph_specs:
            topology_match = self._build_topo_match_from_cgm_spec(spec, component_by_id, relation_by_id, delete_by_id)
            matches.append(topology_match)

        return matches

    def find(self, topology: TopologyResult) -> TopologyMatchingResult:
        component_by_id: dict[str, ComponentWrapper] = {comp.id: comp for comp in topology.components}
        relation_by_id: dict[str, RelationWrapper] = {rel.id: rel for rel in topology.relations}
        delete_by_id: dict[str, TopologyDeleteWrapper] = {dlt.id: dlt for dlt in topology.deletes}

        consistent_graph_matcher = ConsistentGraphMatcher()

        matching_components = self._match_components(topology, consistent_graph_matcher)
        matching_relations = self._match_relations(relation_by_id, matching_components, consistent_graph_matcher)
        matching_deletes = self._match_deletes(topology, consistent_graph_matcher)
        matches = self._match_graphs(consistent_graph_matcher, component_by_id, relation_by_id, delete_by_id)

        return TopologyMatchingResult(
            matches,
            self,
            topology,
            matching_components,
            matching_relations,
            matching_deletes,
            None,
            None
        )
