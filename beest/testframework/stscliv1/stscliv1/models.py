

class ComponentWrapper:
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.name = attributes['name']
        self.type = attributes['type']
        self.tags = attributes['tags']
        self.outgoing_relations = attributes.get('outgoingRelations', [])
        self.incoming_relations = attributes.get('incomingRelations', [])
        self.attributes = attributes

    def __str__(self):
        return 'C[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def __repr__(self):
        return str(self)

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


class TopologyResult:
    def __init__(self, components: list[ComponentWrapper], relations: list[RelationWrapper]):
        self.components = components
        self.relations = relations

    def __str__(self):
        return f"{{{','.join([str(c) for c in self.components])},{','.join([str(r) for r in self.relations])}}}"

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        if isinstance(other, TopologyResult):
            return self.components == other.components and self.relations == other.relations
        return False
