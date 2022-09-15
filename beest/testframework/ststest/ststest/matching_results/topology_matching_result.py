import logging

from stscliv1 import TopologyResult, ComponentWrapper, RelationWrapper
from .matching_result import MatchingResult
from ..primitive_matchers import ComponentMatcher, RelationMatcher
from ..matches import TopologyMatch


class TopologyMatchingResult(MatchingResult):
    def __init__(self,
                 matches: list[TopologyMatch],
                 source: TopologyResult,
                 component_matchers: list[ComponentMatcher],
                 relation_matchers: list[RelationMatcher],
                 component_matches: dict[str, list[ComponentWrapper]],
                 relation_matches: dict[str, list[RelationWrapper]]
                 ):
        super(TopologyMatchingResult, self).__init__()
        self._topology_matches = matches
        self._source = source
        self._component_matches = component_matches
        self.component_matchers = component_matchers
        self.relation_matchers = relation_matchers
        self._relation_matches = relation_matches

    def assert_exact_matches(self, matching_graph_name: str = None) -> list[str]:
        errors = []

        # component matchers
        component_matchers_index = {matcher.id: matcher for matcher in self.component_matchers}
        errors.extend(self._assert_single_match(matches=self._component_matches,
                                                matcher_dict=component_matchers_index))

        # relation matchers
        relation_matchers_index = {matcher.id: matcher for matcher in self.relation_matchers}
        errors.extend(self._assert_single_match(matches=self._relation_matches,
                                                matcher_dict=relation_matchers_index))

        self.dot_graph_drawer.render_debug_dot(topology_matches=self._topology_matches,
                                               source=self._source,
                                               component_matchers=self.component_matchers,
                                               component_matchers_index=component_matchers_index,
                                               component_matches=self._component_matches,
                                               relation_matchers=self.relation_matchers,
                                               relation_matchers_index=relation_matchers_index,
                                               relation_matches=self._relation_matches,
                                               matching_graph_name=matching_graph_name)

        return errors

    def assert_exact_match(self, matching_graph_name=None) -> TopologyMatch:
        if len(self._topology_matches) == 1:
            return self._topology_matches[0]

        errors = self.assert_exact_matches(matching_graph_name)

        error_sep = "\n"
        logging.error(f"desired topology was not matched:\n{error_sep.join(errors)}")
        assert False, f"desired topology was not matched:\n{error_sep.join(errors)}"
