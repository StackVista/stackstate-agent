from stscliv1 import ComponentWrapper, RelationWrapper, TopologyDeleteWrapper, TopologyResult, TopicTopologyResult


def component_fixture(id: int, name: str, outgoing: list[int] = None, incoming=None, tags: list[str] = ()) -> ComponentWrapper:
    return ComponentWrapper({
        'id': id, 'name': name, 'type': 'component',
        'incomingRelations': incoming if incoming is not None else [],
        'outgoingRelations': outgoing if outgoing is not None else [],
        'tags': tags
    })


def relation_fixture(id: int, source: int, target: int, type: str, direction: str = 'ONE_WAY') -> RelationWrapper:
    return RelationWrapper({
        'id': id,
        'source': source,
        'target': target,
        'type': type,
        'dependencyDirection': direction,
    })


def delete_fixture(_: int, delete_id: str) -> TopologyDeleteWrapper:
    return TopologyDeleteWrapper({
        'id': delete_id
    })


class TopologyFixture:
    def __init__(self, definition: str):
        self.elements = {}
        self.components = []
        self.relations = []

        self._setup(definition=definition)

    def _process_component(self,
                           element_definition: str,
                           elem_def_parts: list[str],
                           element_id: int,
                           components: dict[str, ComponentWrapper]) -> dict[str, ComponentWrapper]:
        name = elem_def_parts[0]
        component = component_fixture(element_id, name)
        components[name] = component
        self.elements[element_definition] = component
        self.components.append(component)

        return components

    def _process_relation(self,
                          element_definition: str,
                          elem_def_parts: list[str],
                          element_id: int,
                          components: dict[str, ComponentWrapper]):
        source = components[elem_def_parts[0]]
        relation_type = elem_def_parts[1]
        target = components[elem_def_parts[2]]
        relation = relation_fixture(element_id, source.id, target.id, relation_type)
        self.elements[element_definition] = relation
        self.relations.append(relation)
        source.attributes['outgoingRelations'].append(element_id)
        target.attributes['incomingRelations'].append(element_id)

    def _setup(self, definition: str):
        element_id = 1
        component_ids = {}
        for element_definition in definition.split(","):
            elem_def_parts = element_definition.split(">")
            if len(elem_def_parts) == 1:
                self._process_component(element_definition=element_definition,
                                        elem_def_parts=elem_def_parts,
                                        element_id=element_id,
                                        components=component_ids)
            elif len(elem_def_parts) == 3:
                self._process_relation(element_definition=element_definition,
                                       elem_def_parts=elem_def_parts,
                                       element_id=element_id,
                                       components=component_ids)
            else:
                raise ValueError(
                    f"unrecognized definition `{element_definition}`, should either `some_name` or `some_a>relates>some_b`")

            element_id += 1

    def get(self, definition):
        return self.elements.get(definition)

    def topology(self):
        return TopologyResult(self.components, self.relations)


class TopicTopologyFixture(TopologyFixture):
    def __init__(self, definition: str):
        self.deletes = []

        super(TopicTopologyFixture, self).__init__(definition)

    def _process_component(self,
                           element_definition: str,
                           elem_def_parts: list[str],
                           element_id: int,
                           components: dict[str, ComponentWrapper]) -> dict[str, ComponentWrapper]:
        if element_definition.startswith("del"):
            delete_id = element_definition.split("del ")[1]
            component = delete_fixture(element_id, delete_id)
            self.elements[element_definition] = component
            self.deletes.append(component)
        else:
            name = elem_def_parts[0]
            component = component_fixture(element_id, name)
            components[name] = component
            self.elements[element_definition] = component
            self.components.append(component)

        return components

    def topology(self):
        return TopicTopologyResult(self.components, self.relations, self.deletes)
