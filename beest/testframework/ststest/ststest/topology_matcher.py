from stscliv1 import TopologyResult, ComponentWrapper

from .primitive_matchers import ComponentMatcher, RelationMatcher
from .invariant_search import ConsistentGraphMatcher
from .util import *


class TopologyMatch:
    def __init__(self, components: dict[str, ComponentWrapper], relations: dict[(str, str), RelationWrapper]):
        self._components = components
        self._relations = relations

    def __repr__(self):
        return "Match[\n\t" \
                + "\n\t".join([f"{key}: {comp}" for key, comp in self._components.items()]) \
                + "\n\t" \
                + "\n\t".join([f"{source} > {target}: {comp}" for (source, target), comp in self._relations.items()]) \
                + "\n]"

    def __eq__(self, other):
        if isinstance(other, TopologyMatch):
            return self._components == other._components and self._relations == other._relations
        return False

    def component(self, key: str) -> ComponentWrapper:
        return self._components.get(key)


class TopologyMatchingResult:
    def __init__(self, matches: list[TopologyMatch], errors: list[str]):
        self.errors = errors
        self.matches = matches

    def succeed(self) -> TopologyMatch:
        if len(self.errors) == 0 and len(self.matches) == 1:
            return self.matches[0]
        return None


def get_common_relations(sources: list[ComponentWrapper], targets: list[ComponentWrapper]):
    # TODO consider BOTH_WAY type of relations
    source_relations = set([id for source in sources for id in source.outgoing_relations])
    target_relations = set([id for target in targets for id in target.incoming_relations])
    return list(source_relations & target_relations)


class TopologyMatcher:
    def __init__(self):
        self.components: list[ComponentMatcher] = []
        self.relations: list[RelationMatcher] = []

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
            matching = [comp for comp in topology.components if comp_match.match(comp)]
            if len(matching) == 0:
                add_error(f"component {comp_match} has not been not found")
            else:
                matching_components[comp_match.id] = matching

        # tell CGM that for every virtual node (A) there is a list of possible options (A1..An)
        for key, component_candidates in matching_components.items():
            consistent_graph_matcher.add_choice_of_spec([{key: comp.id} for comp in component_candidates])

        # now we are looking for relations (e.g. A1>B2..Ax>By) that possibly represents a defined relation A>B
        for comp_rel in self.relations:
            source_candidates = matching_components.get(comp_rel.source, [])
            target_candidates = matching_components.get(comp_rel.target, [])
            if len(source_candidates) == 0 or len(target_candidates) == 0:
                add_error(
                    f"relation {comp_rel} has not been found,\n"
                    f"\tsource candidates:\n\t\t{components_short_print(source_candidates)}\n"
                    f"\ttarget candidates:\n\t\t{components_short_print(target_candidates)}")
                continue

            relation_candidate_ids = get_common_relations(source_candidates, target_candidates)
            relation_candidates = [relation_by_id[id] for id in relation_candidate_ids if id in relation_by_id]
            matching_relations = [rel for rel in relation_candidates if comp_rel.match(rel)]
            if len(matching_relations) == 0:
                add_error(
                    f"relation {comp_rel} has not been matched,\n"
                    f"\tsource candidates:\n\t\t{components_short_print(source_candidates)}\n"
                    f"\ttarget candidates:\n\t\t{components_short_print(target_candidates)}\n"
                    f"\tcandidate relations:\n\t\t{relations_short_print(relation_candidates)}"
                )
            else:
                consistent_graph_matcher.add_choice_of_spec([
                    {
                        comp_rel.source: rel.source,
                        comp_rel.target: rel.target,
                        (comp_rel.source, comp_rel.target): rel.id,  # so we can take out a relation id as well
                    }
                    for rel in matching_relations
                ])

        def build_topo_match_from_spec(spec: dict) -> TopologyMatch:
            components = {}
            relations = {}
            for key, id in spec.items():
                if isinstance(key, tuple):
                    relations[key] = relation_by_id[id]
                else:
                    components[key] = component_by_id[id]
            return TopologyMatch(components, relations)

        result_graph_specs = consistent_graph_matcher.get_graphs()
        return TopologyMatchingResult(list(map(build_topo_match_from_spec, result_graph_specs)), errors)


