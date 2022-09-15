from stscliv1 import ComponentWrapper, RelationWrapper, TopologyDeleteWrapper
from ..match_keys import ComponentKey, RelationKey, DeleteKey
from .topology_match import TopologyMatch


class TopicTopologyMatch(TopologyMatch):
    def __init__(self,
                 components: dict[ComponentKey, ComponentWrapper],
                 relations: dict[RelationKey, RelationWrapper],
                 deletes: dict[DeleteKey, TopologyDeleteWrapper]):
        super(TopicTopologyMatch, self).__init__(components, relations)
        self.deletes = deletes

    def __eq__(self, other):
        if isinstance(other, TopicTopologyMatch):
            return super(TopicTopologyMatch, self).__eq__(other) and \
                   self.deletes == other.deletes
        return False

    def delete(self, key: DeleteKey) -> TopologyDeleteWrapper:
        return self.deletes.get(key)

    def __repr__(self):
        parent_repr = super(TopicTopologyMatch, self).__repr__().removesuffix("\n]")

        return f"{parent_repr}" \
               + "\n\t" \
               + "\n\t".join([f"{key}: {dlt}" for key, dlt in self.deletes.items()]) \
               + "\n]"
