from .matcher_builder import TopologyMatcherBuilder
from .topology_matcher import TopologyMatcher
from ..match_keys import RepeatedComponentKey, SingleComponentKey


class RepeatedMatcher(TopologyMatcherBuilder):
    def __init__(self, times: int, parent: 'TopologyMatcher'):
        self.times = times
        self.parent = parent
        self.repeated_components = set()
        self.repeated_elements_flat = set()

    @staticmethod
    def _n_comp_key(id: str, i: int) -> RepeatedComponentKey:
        return id, i

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
