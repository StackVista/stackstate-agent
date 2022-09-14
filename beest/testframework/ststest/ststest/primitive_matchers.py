import re
from stscliv1 import RelationWrapper, ComponentWrapper, TopologyDeleteWrapper
from .match_keys import ComponentKey
from .topology_match import RelationKey

class Matcher:
    @staticmethod
    def matcher_type() -> str:
        return ""


class StringPropertyMatcher(Matcher):
    def __init__(self, key, value):
        self.key = key
        self.value = value

    def __str__(self):
        return f"{self.key}~={self.value}"

    def match(self, value: dict):
        if self.key in value:
            return re.fullmatch(self.value, value[self.key])
        return False

    def matcher_type(self) -> str:
        return "string_property"

class ComponentMatcher(Matcher):
    def __init__(self, id: ComponentKey, props: dict):
        self.id = id
        self.matchers = []
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def __str__(self):
        return f"{self.id}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, component: ComponentWrapper) -> bool:
        for m in self.matchers:
            if not m.match(component.attributes):
                return False
        return True

    def matcher_type(self) -> str:
        return "component"

class RelationMatcher(Matcher):
    def __init__(self, source: ComponentKey, target: ComponentKey, props: dict):
        self.id = (source, target)
        self.source = source
        self.target = target
        self.matchers = []
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def __str__(self):
        return f"{self.source}->{self.target}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, relation: RelationWrapper) -> bool:
        for m in self.matchers:
            if not m.match(relation.attributes):
                return False
        return True

    def matcher_type(self) -> str:
        return "relation"


class DeleteMatcher(Matcher):
    def __init__(self, id: ComponentKey, props: dict):
        self.id = id
        self.matchers = [StringPropertyMatcher('id', id)]
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def __str__(self):
        return f"{self.id}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, delete: TopologyDeleteWrapper) -> bool:
        for m in self.matchers:
            if not m.match(delete.attributes):
                return False
        return True

    def matcher_type(self) -> str:
        return "delete"
