from stscliv1 import TopicTopologyResult, ComponentWrapper, RelationWrapper, TopologyDeleteWrapper

from .topology_matcher import TopologyMatcher
from ..primitive_matchers import DeleteMatcher
from ..invariant_search import ConsistentGraphMatcher
from ..matches import TopicTopologyMatch
from ..topology_matching_result import TopologyMatchingResult, TopicTopologyMatchingResult


class TopicTopologyMatcher(TopologyMatcher):
    def __init__(self):
        super(TopicTopologyMatcher, self).__init__()
        self.delete_matchers: list[DeleteMatcher] = []

    def delete(self, id: str, **kwargs) -> 'TopicTopologyMatcher':
        self.delete_matchers.append(DeleteMatcher(id, kwargs))
        return self

    def _match_deletes(self, topology: TopicTopologyResult,
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
    def _build_topic_topo_match_from_cgm_spec(cgm_spec: dict,
                                              component_by_id: dict[str, ComponentWrapper],
                                              relation_by_id: dict[str, RelationWrapper],
                                              delete_by_id: dict[str, TopologyDeleteWrapper]) -> TopicTopologyMatch:
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

        return TopicTopologyMatch(components, relations, deletes)

    def _match_graphs(self,
                      cgm: ConsistentGraphMatcher,
                      component_by_id: dict[str, ComponentWrapper] = (),
                      relation_by_id: dict[str, RelationWrapper] = (),
                      delete_by_id: dict[str, TopologyDeleteWrapper] = ()) -> list[TopicTopologyMatch]:

        result_graph_specs = cgm.get_graphs()

        matches: list[TopicTopologyMatch] = []
        for spec in result_graph_specs:
            topology_match = self._build_topic_topo_match_from_cgm_spec(
                cgm_spec=spec,
                component_by_id=component_by_id,
                relation_by_id=relation_by_id,
                delete_by_id=delete_by_id
            )
            matches.append(topology_match)

        return matches

    def find(self, topology: TopicTopologyResult) -> TopologyMatchingResult:
        # Take the topology -> TopicTopologyResult and group all data by id
        component_by_id: dict[str, ComponentWrapper] = {comp.id: comp for comp in topology.components}
        relation_by_id: dict[str, RelationWrapper] = {rel.id: rel for rel in topology.relations}
        delete_by_id: dict[str, TopologyDeleteWrapper] = {dlt.id: dlt for dlt in topology.deletes}

        # Initialize the graph matcher to keep track of all the options
        consistent_graph_matcher = ConsistentGraphMatcher()

        # Get all the matching components
        matching_components = self._match_components(topology, consistent_graph_matcher)
        matching_relations = self._match_relations(relation_by_id, matching_components,
                                                   consistent_graph_matcher)
        matching_deletes = self._match_deletes(topology, consistent_graph_matcher)

        matches = self._match_graphs(cgm=consistent_graph_matcher,
                                     component_by_id=component_by_id,
                                     relation_by_id=relation_by_id,
                                     delete_by_id=delete_by_id
                                     )

        return TopicTopologyMatchingResult(
            matches=matches,
            source=topology,
            component_matchers=self.component_matchers,
            relation_matchers=self.relation_matchers,
            delete_matchers=self.delete_matchers,
            component_matches=matching_components,
            relation_matches=matching_relations,
            delete_matches=matching_deletes
        )
