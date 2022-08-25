import json
from typing import Optional
from marshmallow import Schema, fields, EXCLUDE, post_load


class Wrapper:
    def pretty_print(self) -> str:
        pass


class ComponentWrapper(Wrapper):
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.name = attributes['name']
        self.type = attributes['type']
        self.outgoing_relations = attributes.get('outgoingRelations', [])
        self.incoming_relations = attributes.get('incomingRelations', [])
        self.attributes = attributes

    def __str__(self) -> str:
        return 'C[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def pretty_print(self) -> str:
        return f"#{self.id}#[{self.name}]" \
               f"(type={self.type},identifiers={','.join(map(str, self.attributes.get('identifiers', [])))})"

    def __eq__(self, other):
        if isinstance(other, ComponentWrapper):
            return self.attributes == other.attributes
        return False


class RelationWrapper(Wrapper):
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.source = attributes['source']
        self.target = attributes['target']
        self.type = attributes['type']
        self.attributes = attributes

    def __str__(self) -> str:
        return 'R[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def pretty_print(self) -> str:
        return f"#{self.source}->[type={self.type}]->{self.target}"

    def __eq__(self, other) -> bool:
        if isinstance(other, RelationWrapper):
            return self.attributes == other.attributes
        return False


class TopologyStopSnapshotWrapper:
    def __init__(self, offset: int):
        self.offset = offset

    def __str__(self):
        return ']STOP_SNAPSHOT'

    # Topology Stop Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology stop snapshot.
    def __eq__(self, other):
        if isinstance(other, TopologyStopSnapshotWrapper):
            return True
        return False


class TopologyStartSnapshotWrapper:
    def __init__(self, offset: int):
        self.offset = offset

    def __str__(self):
        return 'START_SNAPSHOT['

    # Topology Start Snapshot contains no metadata and is therefore not "unique" within a topic. We can only match on
    # the presence of a topology start snapshot.
    def __eq__(self, other):
        if isinstance(other, TopologyStartSnapshotWrapper):
            return True
        return False


class TopologyDeleteWrapper(Wrapper):
    def __init__(self, attributes: dict):
        self.id = attributes['id']
        self.attributes = attributes

    def pretty_print(self) -> str:
        return f"delete #{self.id}"

    def __str__(self):
        return 'D[' + ','.join([f"{k}={v}" for k, v in self.attributes.items()]) + ']'

    def __eq__(self, other):
        if isinstance(other, TopologyDeleteWrapper):
            return self.attributes == other.attributes
        return False


class TopologyResult:
    def __init__(self,
                 components: list[ComponentWrapper] = [],
                 relations: list[RelationWrapper] = []):
        self.components = components
        self.relations = relations

    def component(self, component: ComponentWrapper):
        self.components.append(component)

    def relation(self, relation: RelationWrapper):
        self.relations.append(relation)

    def __str__(self):
        components = ','.join([str(c) for c in self.components])
        relations = ','.join([str(r) for r in self.relations])

        return f"{{{components},{relations}}}"

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        if isinstance(other, TopologyResult):
            return self.components == other.components and \
                   self.relations == other.relations
        return False


class TopicTopologyResult(TopologyResult):
    def __init__(self,
                 components: list[ComponentWrapper] = [],
                 relations: list[RelationWrapper] = [],
                 deletes: list[TopologyDeleteWrapper] = []):
        super(TopicTopologyResult, self).__init__(components, relations)
        self.deletes = deletes

    def delete(self, delete: TopologyDeleteWrapper):
        self.deletes.append(delete)

    def __str__(self):
        parent_str = super(TopicTopologyResult, self).__str__()
        deletes = ','.join([str(r) for r in self.deletes])

        return f"{{{parent_str},{deletes}}}"

    def __eq__(self, other):
        if isinstance(other, TopicTopologyResult):
            return super(TopicTopologyResult, self).__eq__(other) and \
                   self.deletes == other.deletes
        return False


class TopologySnapshotResult:
    def __init__(self):
        self._start_snapshot: Optional[TopologyStartSnapshotWrapper] = None
        self._topic_topology = TopicTopologyResult([], [], [])
        self._stop_snapshot: Optional[TopologyStopSnapshotWrapper] = None

    def start_snapshot(self, offset):
        self._start_snapshot = TopologyStartSnapshotWrapper(offset)

    def get_start_snapshot(self) -> Optional[TopologyStartSnapshotWrapper]:
        return self._start_snapshot

    def component(self, component: ComponentWrapper):
        self._topic_topology.component(component)

    def relation(self, relation: RelationWrapper):
        self._topic_topology.relation(relation)

    def delete(self, delete: TopologyDeleteWrapper):
        self._topic_topology.delete(delete)

    def stop_snapshot(self, offset):
        self._stop_snapshot = TopologyStopSnapshotWrapper(offset)

    def get_stop_snapshot(self) -> Optional[TopologyStopSnapshotWrapper]:
        return self._stop_snapshot

    def __str__(self):
        start_snapshot = str(self._start_snapshot) if self._start_snapshot else ''
        components = ','.join([str(c) for c in self._topic_topology.components])
        relations = ','.join([str(r) for r in self._topic_topology.relations])
        deletes = ','.join([str(r) for r in self._topic_topology.deletes])
        stop_snapshot = str(self._stop_snapshot) if self._stop_snapshot else ''

        return f"{{{start_snapshot}{components},{relations},{deletes}{stop_snapshot}}}"

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        if isinstance(other, TopologySnapshotResult):
            return self._start_snapshot == other._start_snapshot and \
                   self._topic_topology == other._topic_topology and \
                   self._stop_snapshot == other.stop_snapshot
        return False


class TopologyComponent:
    def __init__(self, external_id: str, type_name: str, data: dict):
        self.externalId = external_id
        self.typeName = type_name
        self.data = data

    def wrap(self) -> ComponentWrapper:
        return ComponentWrapper({
            'id': self.externalId,
            'name': self.data.get('name', self.externalId),
            'type': self.typeName,
            **self.data
        })


class TopologyComponentSchema(Schema):
    class Meta:
        unknown = EXCLUDE

    externalId = fields.String()
    typeName = fields.String()
    data = fields.String()

    @post_load
    def new_model(self, data, **kwargs):
        return TopologyComponent(
            external_id=data.get('externalId'),
            type_name=data.get('typeName'),
            data=json.loads(data.get('data'))
        )


class TopologyRelation:
    def __init__(self, external_id: str, type_name: str, source_id: str, target_id: str, data: dict):
        self.externalId = external_id
        self.typeName = type_name
        self.source_id = source_id
        self.target_id = target_id
        self.data = data

    def wrap(self) -> RelationWrapper:
        return RelationWrapper({
            'id': self.externalId,
            'source': self.source_id,
            'target': self.target_id,
            'type': self.typeName,
            **self.data
        })


class TopologyRelationSchema(Schema):
    class Meta:
        unknown = EXCLUDE

    externalId = fields.String()
    typeName = fields.String()
    sourceId = fields.String()
    targetId = fields.String()
    data = fields.String()

    @post_load
    def new_model(self, data, **kwargs):
        return TopologyRelation(
            external_id=data.get('externalId'),
            type_name=data.get('typeName'),
            source_id=data.get('sourceId'),
            target_id=data.get('targetId'),
            data=json.loads(data.get('data'))
        )


class TopologyDelete:
    def __init__(self, external_id: str):
        self.external_id = external_id

    def wrap(self) -> TopologyDeleteWrapper:
        return TopologyDeleteWrapper({
            'id': self.external_id,
            **vars(self)
        })


class TopologyDeleteSchema(Schema):
    class Meta:
        unknown = EXCLUDE

    externalId = fields.String()

    @post_load
    def new_model(self, data, **kwargs):
        return TopologyDelete(external_id=data.get('externalId'))


class Payload:
    def __init__(self,
                 topology_start_snapshot: Optional[dict],
                 topology_component: Optional[TopologyComponent],
                 topology_relation: Optional[TopologyRelation],
                 topology_delete: Optional[TopologyDelete],
                 topology_stop_snapshot: Optional[dict]):
        self.topology_start_snapshot = topology_start_snapshot
        self.topology_component = topology_component
        self.topology_relation = topology_relation
        self.topology_delete = topology_delete
        self.topology_stop_snapshot = topology_stop_snapshot


class PayloadSchema(Schema):
    class Meta:
        unknown = EXCLUDE

    TopologyStartSnapshot = fields.Dict(keys=fields.String(), values=fields.String())
    TopologyComponent = fields.Nested(TopologyComponentSchema())
    TopologyRelation = fields.Nested(TopologyRelationSchema())
    TopologyDelete = fields.Nested(TopologyDeleteSchema())
    TopologyStopSnapshot = fields.Dict(keys=fields.String(), values=fields.String())

    @post_load
    def new_model(self, data, **kwargs):
        return Payload(
            topology_start_snapshot=data.get('TopologyStartSnapshot'),
            topology_component=data.get('TopologyComponent'),
            topology_relation=data.get('TopologyRelation'),
            topology_delete=data.get('TopologyDelete'),
            topology_stop_snapshot=data.get('TopologyStopSnapshot')
        )


class TopologyElement:
    def __init__(self, payload: Payload):
        self.payload = payload


class TopologyElementSchema(Schema):
    collectionTimestamp = fields.Number()
    payload = fields.Nested(PayloadSchema())
    ingestionTimestamp = fields.Number()

    @post_load
    def new_model(self, data, **kwargs):
        return TopologyElement(
            payload=data.get('payload')
        )


class Message:
    def __init__(self, topology_element: TopologyElement):
        self.topology_element = topology_element


class MessageSchema(Schema):
    TopologyElement = fields.Nested(TopologyElementSchema())

    @post_load
    def new_model(self, data, **kwargs):
        return Message(
            topology_element=data.get('TopologyElement')
        )


class TopicMessage:
    def __init__(self, message: Message, offset: int):
        self.message = message
        self.offset = offset


class TopicMessageSchema(Schema):
    partition = fields.Number()
    offset = fields.Number()
    message = fields.Nested(MessageSchema())

    @post_load
    def new_model(self, data, **kwargs):
        return TopicMessage(
            message=data.get('message'),
            offset=data.get('offset')
        )


class TopicAPIResponse:
    def __init__(self, messages: list[TopicMessage]):
        self.messages = messages


class TopicAPIResponseSchema(Schema):
    messages = fields.List(fields.Nested(TopicMessageSchema()))

    @post_load
    def new_model(self, data, **kwargs):
        return TopicAPIResponse(**data)
