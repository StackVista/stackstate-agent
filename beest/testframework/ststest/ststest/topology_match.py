from typing import Tuple, Union

from stscliv1 import ComponentWrapper, RelationWrapper
from .match_keys import ComponentKey, RelationKey, SingleComponentKey


class TopologyMatch:
    def __init__(self, components: dict[ComponentKey, ComponentWrapper], relations: dict[RelationKey, RelationWrapper]):
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

    def component(self, key: SingleComponentKey) -> ComponentWrapper:
        return self._components.get(key)

    def repeated_components(self, key: SingleComponentKey) -> list[ComponentWrapper]:
        return [comp for (ckey, comp) in self._components.items() if isinstance(ckey, tuple) and ckey[0] == key]

    def has_component(self, id: int) -> bool:
        return next((True for comp in self._components.values() if comp.id == id), False)

    def has_relation(self, id: int) -> bool:
        return next((True for rel in self._relations.values() if rel.id == id), False)
