import re
from stscliv1 import RelationWrapper, ComponentWrapper


class StringPropertyMatcher:
    def __init__(self, key, value):
        self.key = key
        self.value = value

    def __str__(self):
        return f"{self.key}~={self.value}"

    def match(self, value: dict):
        if self.key in value:
            return re.fullmatch(self.value, value[self.key])
        return False


class ComponentMatcher:
    def __init__(self, id: str, props: dict):
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


class RelationMatcher:
    def __init__(self, source: str, target: str, props: dict):
        self.source = source
        self.target = target
        self.matchers = []
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def id(self):
        return f"{self.source}_TO_{self.target}"

    def __str__(self):
        return f"{self.source}->{self.target}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, relation: RelationWrapper) -> bool:
        for m in self.matchers:
            if not m.match(relation.attributes):
                return False
        return True
