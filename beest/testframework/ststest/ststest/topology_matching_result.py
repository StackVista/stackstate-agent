import logging

from .primitive_matchers import Matcher, ComponentMatcher, RelationMatcher, DeleteMatcher
from .topology_match import TopologyMatch, TopicTopologyMatch
from .dot_graph_drawer import DotGraphDrawer
from stscliv1 import Wrapper, TopologyResult, ComponentWrapper, RelationWrapper, TopologyDeleteWrapper


class MatchingResult:
    def __init__(self,
                 matches: list[TopologyMatch],
                 source: TopologyResult,
                 ):
        self._topology_matches = matches
        self._source = source
        self.dot_graph_drawer = DotGraphDrawer()

    @staticmethod
    def _assert_single_match(matches: dict[str, list[Wrapper]], matcher_dict: dict[str, Matcher]) -> list[str]:
        errors = []
        delimiter = "\n\t\t"

        for key, items in matches.items():
            matcher = matcher_dict[key]
            if len(items) == 0:
                errors.append(f"\t{matcher.matcher_type()} {matcher} was not found")
            elif len(items) > 1:
                matches_str = delimiter.join(map(lambda item: item.pretty_print(), items))
                errors.append(f"\tmultiple matches for {matcher.matcher_type()} {matcher}:"
                              f"{delimiter}{matches_str}")

        return errors

    def assert_exact_matches(self, matching_graph_name=None, matching_graph_upload=True) -> list[str]:
        pass

    def assert_exact_match(self, matching_graph_name=None, matching_graph_upload=True, strict=True) -> TopologyMatch:
        if len(self._topology_matches) == 1:
            return self._topology_matches[0]
        elif len(self._topology_matches) > 1 and not strict:
            return self._topology_matches[0]

        errors = self.assert_exact_matches(matching_graph_name, matching_graph_upload)

        error_sep = "\n"
        logging.error(f"desired topology was not matched:\n{error_sep.join(errors)}")
        assert False, f"desired topology was not matched:\n{error_sep.join(errors)}"


class TopologyMatchingResult(MatchingResult):
    def __init__(self,
                 matches: list[TopologyMatch],
                 source: TopologyResult,
                 component_matchers: list[ComponentMatcher],
                 relation_matchers: list[RelationMatcher],
                 component_matches: dict[str, list[ComponentWrapper]],
                 relation_matches: dict[str, list[RelationWrapper]]
                 ):
        super(TopologyMatchingResult, self).__init__(matches, source)
        self._component_matches = component_matches
        self.component_matchers = component_matchers
        self.relation_matchers = relation_matchers
        self._relation_matches = relation_matches

    def assert_exact_matches(self, matching_graph_name: str = None, matching_graph_upload: bool = True) -> list[str]:
        errors = []

        # component matchers
        comp_matchers = {matcher.id: matcher for matcher in self.component_matchers}
        errors.extend(self._assert_single_match(matches=self._component_matches,
                                                matcher_dict=comp_matchers))

        # relation matchers
        rel_matchers = {matcher.id(): matcher for matcher in self.relation_matchers}
        errors.extend(self._assert_single_match(matches=self._relation_matches,
                                                matcher_dict=rel_matchers))

        self.dot_graph_drawer.render_debug_dot(topology_matches=self._topology_matches,
                                               source=self._source,
                                               component_matchers=self.component_matchers,
                                               component_matches=self._component_matches,
                                               relation_matchers=self.relation_matchers,
                                               relation_matches=self._relation_matches,
                                               matching_graph_name=matching_graph_name,
                                               generate_diagram_url=matching_graph_upload)

        return errors


class TopicTopologyMatchingResult(TopologyMatchingResult):
    def __init__(self,
                 matches: list[TopicTopologyMatch],
                 source: TopologyResult,
                 component_matchers: list[ComponentMatcher],
                 relation_matchers: list[RelationMatcher],
                 delete_matchers: list[DeleteMatcher],
                 component_matches: dict[str, list[ComponentWrapper]],
                 relation_matches: dict[str, list[RelationWrapper]],
                 delete_matches: dict[str, list[TopologyDeleteWrapper]]
                 ):
        super(TopicTopologyMatchingResult, self).__init__(matches=matches,
                                                          source=source,
                                                          component_matchers=component_matchers,
                                                          relation_matchers=relation_matchers,
                                                          component_matches=component_matches,
                                                          relation_matches=relation_matches)
        self.delete_matchers = delete_matchers
        self._delete_matches = delete_matches

    def assert_exact_matches(self, matching_graph_name: str = None, matching_graph_upload: bool = True) -> list[str]:
        errors = []

        # component matchers
        comp_matchers = {matcher.id: matcher for matcher in self.component_matchers}
        errors.extend(self._assert_single_match(matches=self._component_matches,
                                                matcher_dict=comp_matchers))

        # relation matchers
        rel_matchers = {matcher.id(): matcher for matcher in self.relation_matchers}
        errors.extend(self._assert_single_match(matches=self._relation_matches,
                                                matcher_dict=rel_matchers))

        # delete matchers
        del_matchers = {matcher.id: matcher for matcher in self.delete_matchers}
        errors.extend(self._assert_single_match(matches=self._delete_matches,
                                                matcher_dict=del_matchers))

        self.dot_graph_drawer.render_debug_dot(topology_matches=self._topology_matches,
                                               source=self._source,
                                               component_matchers=self.component_matchers,
                                               component_matches=self._component_matches,
                                               relation_matchers=self.relation_matchers,
                                               relation_matches=self._relation_matches,
                                               matching_graph_name=matching_graph_name,
                                               generate_diagram_url=matching_graph_upload)

        return errors
