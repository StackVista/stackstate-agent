import json
from typing import Optional
from marshmallow import Schema, fields, EXCLUDE, post_load


class TopologyComponent:
    def __init__(self, external_id: str, type_name: str, data: dict):
        self.externalId = external_id
        self.typeName = type_name
        self.data = data


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


class TopologyDeleteSchema(Schema):
    class Meta:
        unknown = EXCLUDE

    externalId = fields.String()

    @post_load
    def new_model(self, data, **kwargs):
        return TopologyDelete(external_id=data.get('external_id'))


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


class TopologySnapshotResult:
    def __init__(self):
        self._start_snapshot: Optional[TopologyStartSnapshotWrapper] = None
        self._components: list[ComponentWrapper] = []
        self._relations: list[RelationWrapper] = []
        self._deletes: list[TopologyDeleteWrapper] = []
        self._stop_snapshot: Optional[TopologyStopSnapshotWrapper] = None

    def start_snapshot(self, offset):
        self._start_snapshot = TopologyStartSnapshotWrapper(offset)

    def component(self, component: ComponentWrapper):
        self._components.append(component)

    def relation(self, relation: RelationWrapper):
        self._relations.append(relation)

    def delete(self, delete: TopologyDeleteWrapper):
        self._deletes.append(delete)

    def stop_snapshot(self, offset):
        self._stop_snapshot = TopologyStopSnapshotWrapper(offset)

    def __str__(self):
        start_snapshot = str(self._start_snapshot) if self._start_snapshot else ''
        components = ','.join([str(c) for c in self._components])
        relations = ','.join([str(r) for r in self._relations])
        deletes = ','.join([str(r) for r in self._deletes])
        stop_snapshot = str(self._stop_snapshot) if self._stop_snapshot else ''

        return f"{{{start_snapshot}{components},{relations},{deletes}{stop_snapshot}}}"

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        if isinstance(other, TopologyResult):
            return self._start_snapshot == other.start_snapshot and \
                   self._components == other.components and \
                   self._relations == other.relations and \
                   self._deletes == other.deletes and \
                   self._stop_snapshot == other.stop_snapshot
        return False
