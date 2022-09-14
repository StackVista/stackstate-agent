from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper
from typing import Callable
from .graph_matcher import GraphMatcher
from .matcher_builder import TopologyMatcherBuilder
from repeated_matcher import RepeatedMatcher
from .common import get_common_relations, filter_out_repeated_specs
from ..primitive_matchers import ComponentMatcher, RelationMatcher
from ..invariant_search import ConsistentGraphMatcher
from ..matches import TopologyMatch
from ..topology_matching_result import TopologyMatchingResult
from ..match_keys import ComponentKey


class TopologyMatcher(GraphMatcher):

    def __init__(self):
        super(TopologyMatcher, self).__init__()
        self.component_matchers: list[ComponentMatcher] = []
        self.relation_matchers: list[RelationMatcher] = []
        self._ambiguous_elements: set = set()

    def repeated(self, times: int, define_submatch: Callable[[TopologyMatcherBuilder], TopologyMatcherBuilder]):
        repeated_matcher = RepeatedMatcher(times, self)
        define_submatch(repeated_matcher)
        self._ambiguous_elements = self._ambiguous_elements | repeated_matcher.repeated_elements_flat
        return self

    def component(self, id: ComponentKey, **kwargs) -> 'TopologyMatcher':
        self.component_matchers.append(ComponentMatcher(id, kwargs))
        return self

    def one_way_direction(self, source: ComponentKey, target: ComponentKey, **kwargs) -> 'TopologyMatcher':
        source_found = False
        target_found = False
        for comp in self.component_matchers:
            if comp.id == source:
                source_found = True
            if comp.id == target:
                target_found = True
        if not source_found:
            raise KeyError(
                f"source `{source}` have not been found, use .component('{source}') to define a component "
                f"before defining a relation")
        if not target_found:
            raise KeyError(
                f"target `{target}` have not been found, use .component('{target}') to define a component "
                f"before defining a relation")

        kwargs['dependencyDirection'] = 'ONE_WAY'
        self.relation_matchers.append(RelationMatcher(source, target, kwargs))
        return self

    def _match_components(self, topology: TopologyResult,
                          cgm: ConsistentGraphMatcher) -> dict[str, list[ComponentWrapper]]:

        # find all matching components and group them by virtual node (id) from a pattern
        matching_components: dict[str, list[ComponentWrapper]] = {}
        for comp_match in self.component_matchers:
            matching_components[comp_match.id] = [comp for comp in topology.components if
                                                  comp_match.match(comp)]

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
            matching_relations[comp_rel.id] = matching
            cgm.add_choice_of_spec([
                {
                    comp_rel.source: rel.source,
                    comp_rel.target: rel.target,
                    comp_rel.id: rel.id,
                }
                for rel in matching
            ])

        return matching_relations

    @staticmethod
    def _build_topo_match_from_cgm_spec(cgm_spec: dict,
                                        component_by_id: dict[str, ComponentWrapper],
                                        relation_by_id: dict[str, RelationWrapper]) -> TopologyMatch:
        components = {}
        relations = {}
        for key, spec_id in cgm_spec.items():
            if spec_id in relation_by_id:
                relations[key] = relation_by_id[spec_id]
            else:
                components[key] = component_by_id[spec_id]

        return TopologyMatch(components, relations)

    def _match_graphs(self,
                      cgm: ConsistentGraphMatcher,
                      component_by_id: dict[str, ComponentWrapper],
                      relation_by_id: dict[str, RelationWrapper]) -> list[TopologyMatch]:

        result_graph_specs = cgm.get_graphs()
        distinct_graph_specs = filter_out_repeated_specs(result_graph_specs, self._ambiguous_elements)

        matches: list[TopologyMatch] = []
        for spec in distinct_graph_specs:
            topology_match = self._build_topo_match_from_cgm_spec(
                cgm_spec=spec,
                component_by_id=component_by_id,
                relation_by_id=relation_by_id
            )
            matches.append(topology_match)

        return matches

    def find(self, topology: TopologyResult) -> TopologyMatchingResult:
        # Take the topology -> TopologyResult and group all data by id
        component_by_id: dict[str, ComponentWrapper] = {comp.id: comp for comp in topology.components}
        relation_by_id: dict[str, RelationWrapper] = {rel.id: rel for rel in topology.relations}

        # Initialize the graph matcher to keep track of all the options
        consistent_graph_matcher = ConsistentGraphMatcher()

        # Get all the matching components
        matching_components = self._match_components(topology, consistent_graph_matcher)
        matching_relations = self._match_relations(relation_by_id, matching_components,
                                                   consistent_graph_matcher)

        matches = self._match_graphs(cgm=consistent_graph_matcher,
                                     component_by_id=component_by_id,
                                     relation_by_id=relation_by_id
                                     )

        return TopologyMatchingResult(
            matches=matches,
            source=topology,
            component_matchers=self.component_matchers,
            relation_matchers=self.relation_matchers,
            component_matches=matching_components,
            relation_matches=matching_relations
        )
