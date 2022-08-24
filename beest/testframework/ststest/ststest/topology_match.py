from stscliv1 import ComponentWrapper, RelationWrapper, TopologyDeleteWrapper


class TopologyMatch:
    def __init__(self,
                 components: dict[str, ComponentWrapper],
                 relations: dict[str, RelationWrapper]):
        self._components = components
        self._relations = relations

    def __repr__(self):
        components = "\n".join([f"{key}: {comp}" for key, comp in self._components.items()])
        relations = "\n".join([f"{rel.source} > {rel.target}: {rel}" for _, rel in self._relations.items()])

        return f"Match" \
               f"\n[Components]\n" \
               f"{components}" \
               f"\n[Relations]\n" \
               f"{relations}" \
               "\n"

    def __eq__(self, other):
        if isinstance(other, TopologyMatch):
            return self._components == other._components and \
                   self._relations == other._relations
        return False

    def component(self, key: str) -> ComponentWrapper:
        return self._components.get(key)

    def has_component(self, id: int) -> bool:
        return next((True for comp in self._components.values() if comp.id == id), False)

    def has_relation(self, id: int) -> bool:
        return next((True for rel in self._relations.values() if rel.id == id), False)


class TopicTopologyMatch(TopologyMatch):
    def __init__(self,
                 components: dict[str, ComponentWrapper],
                 relations: dict[str, RelationWrapper],
                 deletes: dict[str, TopologyDeleteWrapper]):
        super(TopicTopologyMatch, self).__init__(components, relations)
        self._deletes = deletes

    def __eq__(self, other):
        if isinstance(other, TopicTopologyMatch):
            return super(TopicTopologyMatch, self).__eq__(other) and \
                   self._deletes == other._deletes
        return False

    def delete(self, key) -> TopologyDeleteWrapper:
        return self._deletes.get(key)

    def __repr__(self):
        parent_repr = super(TopicTopologyMatch, self).__repr__()
        deletes = "\n".join([f"{key}: {dlt}" for key, dlt in self._deletes.items()])

        return f"{parent_repr}" \
               f"\n[Deletes]\n" \
               f"{deletes}" \
               "\n"
