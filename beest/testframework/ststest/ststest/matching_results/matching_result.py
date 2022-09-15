from stscliv1 import Wrapper
from ..dot_graph_drawer import DotGraphDrawer
from ..primitive_matchers import Matcher


class MatchingResult:
    def __init__(self):
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

    def assert_exact_matches(self, matching_graph_name=None) -> list[str]:
        pass
