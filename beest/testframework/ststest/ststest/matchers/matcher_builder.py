from abc import abstractmethod


class TopologyMatcherBuilder:
    @abstractmethod
    def component(self, id, **kwargs) -> 'TopologyMatcherBuilder':
        raise NotImplementedError()

    @abstractmethod
    def one_way_direction(self, source, target, **kwargs) -> 'TopologyMatcherBuilder':
        raise NotImplementedError()
