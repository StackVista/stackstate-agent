from typing import Optional


class ComponentWrapper:
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.name = attributes['name']
        self.type = attributes['type']
        self.outgoing_relations = attributes.get('outgoingRelations', [])
        self.incoming_relations = attributes.get('incomingRelations', [])
        self.attributes = attributes

    def __str__(self):
        return 'C[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def __eq__(self, other):
        if isinstance(other, ComponentWrapper):
            return self.attributes == other.attributes
        return False


class RelationWrapper:
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.source = attributes['source']
        self.target = attributes['target']
        self.type = attributes['type']
        self.attributes = attributes

    def __str__(self):
        return 'R[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def __eq__(self, other):
        if isinstance(other, RelationWrapper):
            return self.attributes == other.attributes
        return False


class TopologyStopSnapshotWrapper:
    def __str__(self):
        return ']STOP_SNAPSHOT'

    # Topology Stop Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology stop snapshot.
    def __eq__(self, other):
        if isinstance(other, TopologyStopSnapshotWrapper):
            return True
        return False


class TopologyStartSnapshotWrapper:
    def __str__(self):
        return 'START_SNAPSHOT['

    # Topology Start Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology start snapshot.
    def __eq__(self, other):
        if isinstance(other, TopologyStartSnapshotWrapper):
            return True
        return False


class TopologyDeleteWrapper:
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.attributes = attributes

    def __str__(self):
        return 'D[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def __eq__(self, other):
        if isinstance(other, TopologyDeleteWrapper):
            return self.attributes == other.attributes
        return False


class TopologyResult:
    def __init__(self, components: list[ComponentWrapper] = (), relations: list[RelationWrapper] = (),
                 deletes: list[TopologyDeleteWrapper] = (), start_snapshot: Optional[TopologyStartSnapshotWrapper] = None,
                 stop_snapshot: Optional[TopologyStopSnapshotWrapper] = None):
        self.components = components
        self.relations = relations
        self.deletes = deletes
        self.start_snapshot = start_snapshot
        self.stop_snapshot = stop_snapshot

    def __str__(self):
        components = ','.join([str(c) for c in self.components])
        relations = ','.join([str(r) for r in self.relations])
        deletes = ','.join([str(r) for r in self.deletes])
        start_snapshot = str(self.start_snapshot) if self.start_snapshot else ''
        stop_snapshot = str(self.stop_snapshot) if self.stop_snapshot else ''

        return f"{{{start_snapshot}{components},{relations},{deletes}{stop_snapshot}}}"

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        if isinstance(other, TopologyResult):
            return self.components == other.components and \
                   self.relations == other.relations and \
                   self.deletes == other.deletes and \
                   self.start_snapshot == other.start_snapshot and \
                   self.stop_snapshot == other.stop_snapshot
        return False
