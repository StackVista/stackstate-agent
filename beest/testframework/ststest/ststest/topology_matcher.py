from abc import abstractmethod
from typing import Callable

from stscliv1 import TopologyResult, ComponentWrapper

from .primitive_matchers import ComponentMatcher, RelationMatcher
from .invariant_search import ConsistentGraphMatcher
from .topology_match import TopologyMatch
from .match_keys import RepeatedComponentKey, ComponentKey, SingleComponentKey
from .topology_matching_result import TopologyMatchingResult


def get_common_relations(sources: list[ComponentWrapper], targets: list[ComponentWrapper]):
    # TODO consider BOTH_WAY type of relations
    source_relations = set([id for source in sources for id in source.outgoing_relations])
    target_relations = set([id for target in targets for id in target.incoming_relations])
    return list(source_relations & target_relations)


def filter_out_repeated_specs(specs: list[dict], ambiguous_elements: set[str]):
    def two_are_equivalent(spec1: dict, spec2: dict):
        for key in (set(spec1.keys()) | set(spec2.keys())):
            if key not in ambiguous_elements and spec1[key] != spec2[key]:
                return False
        return True

    distinct_specs = []
    for spec in specs:
        spec_is_distinct = True
        for dspec in distinct_specs:
            if two_are_equivalent(spec, dspec):
                spec_is_distinct = False
                break
        if spec_is_distinct:
            distinct_specs.append(spec)

    return distinct_specs


class TopologyMatcherBuilder:
    @abstractmethod
    def component(self, id, **kwargs) -> 'TopologyMatcherBuilder':
        raise NotImplementedError()

    @abstractmethod
    def one_way_direction(self, source, target, **kwargs) -> 'TopologyMatcherBuilder':
        raise NotImplementedError()


class RepeatedMatcher(TopologyMatcherBuilder):
    def __init__(self, times: int, parent: 'TopologyMatcher'):
        self.times = times
        self.parent = parent
        self.repeated_components = set()
        self.repeated_elements_flat = set()

    @staticmethod
    def _n_comp_key(id: str, i: int) -> RepeatedComponentKey:
        return (id, i)

    def component(self, id: SingleComponentKey, **kwargs) -> 'RepeatedMatcher':
        self.repeated_components.add(id)
        for i in range(0, self.times):
            idN = self._n_comp_key(id, i)
            self.parent.component(idN, **kwargs)
            self.repeated_elements_flat.add(idN)
        return self

    def one_way_direction(self, source: SingleComponentKey, target: SingleComponentKey, **kwargs) -> 'RepeatedMatcher':
        for i in range(0, self.times):
            source_i = self._n_comp_key(source, i) if source in self.repeated_components else source
            target_i = self._n_comp_key(target, i) if target in self.repeated_components else target
            self.parent.one_way_direction(source_i, target_i, **kwargs)
            self.repeated_elements_flat.add((source_i, target_i))
        return self


class TopologyMatcher(TopologyMatcherBuilder):
    def __init__(self):
        self._components: list[ComponentMatcher] = []
        self._relations: list[RelationMatcher] = []
        self._ambiguous_elements: set = set()

    def repeated(self, times: int, define_submatch: Callable[[TopologyMatcherBuilder], TopologyMatcherBuilder]):
        repeated_matcher = RepeatedMatcher(times, self)
        define_submatch(repeated_matcher)
        self._ambiguous_elements = self._ambiguous_elements | repeated_matcher.repeated_elements_flat
        return self

    def component(self, id: ComponentKey, **kwargs) -> 'TopologyMatcher':
        self._components.append(ComponentMatcher(id, kwargs))
        return self

    def one_way_direction(self, source: ComponentKey, target: ComponentKey, **kwargs) -> 'TopologyMatcher':
        source_found = False
        target_found = False
        for comp in self._components:
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
        self._relations.append(RelationMatcher(source, target, kwargs))
        return self

    def find(self, topology: TopologyResult) -> TopologyMatchingResult:
        component_by_id = {comp.id: comp for comp in topology.components}
        relation_by_id = {rel.id: rel for rel in topology.relations}

        consistent_graph_matcher = ConsistentGraphMatcher()

        # find all matching components and group them by virtual node (id) from a pattern
        matching_components: dict[str, list[ComponentWrapper]] = {}
        for comp_match in self._components:
            matching_components[comp_match.id] = [comp for comp in topology.components if comp_match.match(comp)]

        # tell CGM that for every virtual node (A) there is a list of possible options (A1..An)
        for key, component_candidates in matching_components.items():
            consistent_graph_matcher.add_choice_of_spec([{key: comp.id} for comp in component_candidates])

        # now we are looking for relations (e.g. A1>B2..Ax>By) that possibly represents a defined relation A>B
        matching_relations = {}
        for comp_rel in self._relations:
            source_candidates = matching_components.get(comp_rel.source, [])
            target_candidates = matching_components.get(comp_rel.target, [])
            relation_candidate_ids = get_common_relations(source_candidates, target_candidates)
            relation_candidates = [relation_by_id[id] for id in relation_candidate_ids if id in relation_by_id]
            matching = [rel for rel in relation_candidates if comp_rel.match(rel)]
            matching_relations[comp_rel.id] = matching
            consistent_graph_matcher.add_choice_of_spec([
                {
                    comp_rel.source: rel.source,
                    comp_rel.target: rel.target,
                    comp_rel.id: rel.id,
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
        distinct_graph_specs = filter_out_repeated_specs(result_graph_specs, self._ambiguous_elements)
        return TopologyMatchingResult(
            list(map(build_topo_match_from_spec, distinct_graph_specs)),
            self,
            topology,
            matching_components,
            matching_relations,
        )
