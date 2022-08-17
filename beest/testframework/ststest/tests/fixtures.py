from stscliv1 import ComponentWrapper, RelationWrapper, TopologyDeleteWrapper, TopologyResult


def component_fixture(id: int, name: str, outgoing: list[int] = None, incoming=None) -> ComponentWrapper:
    return ComponentWrapper({
        'id': id, 'name': name, 'type': 'component',
        'incomingRelations': incoming if incoming is not None else [],
        'outgoingRelations': outgoing if outgoing is not None else [],
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
        self.deletes = []

        element_id = 1
        component_ids = {}
        for element_definition in definition.split(","):
            elem_def_parts = element_definition.split(">")
            if len(elem_def_parts) == 1:
                if element_definition.startswith("del"):
                    delete_id = element_definition.split("del ")[1]
                    component = delete_fixture(element_id, delete_id)
                    self.elements[element_definition] = component
                    self.deletes.append(component)
                else:
                    name = elem_def_parts[0]
                    component = component_fixture(element_id, name)
                    component_ids[name] = component
                    self.elements[element_definition] = component
                    self.components.append(component)
            elif len(elem_def_parts) == 3:
                source = component_ids[elem_def_parts[0]]
                type = elem_def_parts[1]
                target = component_ids[elem_def_parts[2]]
                relation = relation_fixture(element_id, source.id, target.id, type)
                self.elements[element_definition] = relation
                self.relations.append(relation)
                source.attributes['outgoingRelations'].append(element_id)
                target.attributes['incomingRelations'].append(element_id)
            else:
                raise ValueError(
                    f"unrecognized definition `{element_definition}`, should either `some_name` or `some_a>relates>some_b`")
            element_id += 1

    def get(self, definition):
        return self.elements.get(definition)

    def topology(self):
        return TopologyResult(self.components, self.relations, self.deletes)
