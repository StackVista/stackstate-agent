import re
from stscliv1 import RelationWrapper, ComponentWrapper, TopologyDeleteWrapper, TopologyStartSnapshotWrapper, \
    TopologyStopSnapshotWrapper


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

    def matcher_type(self) -> str:
        return "component"


class RelationMatcher(Matcher):
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

    def matcher_type(self) -> str:
        return "relation"


class DeleteMatcher(Matcher):
    def __init__(self, id: str, props: dict):
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


class StartSnapshotMatcher(Matcher):
    def __init__(self, id: str):
        self.id = id

    def __str__(self):
        return 'START_SNAPSHOT['

    # Topology Start Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology start snapshot.
    def match(self, start_snapshot: TopologyStartSnapshotWrapper) -> bool:
        if start_snapshot:
            return True
        return False

    def matcher_type(self) -> str:
        return "start_snapshot"


class StopSnapshotMatcher(Matcher):
    def __init__(self, id: str):
        self.id = id

    def __str__(self):
        return ']STOP_SNAPSHOT'

    # Topology Stop Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology stop snapshot.
    def match(self, stop_snapshot: TopologyDeleteWrapper) -> bool:
        if stop_snapshot:
            return True
        return False

    def matcher_type(self) -> str:
        return "stop_snapshot"
