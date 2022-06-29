from stscliv1 import ComponentWrapper, RelationWrapper


def component_pretty_short(comp: ComponentWrapper):
    return f"#{comp.id}#[{comp.name}](type={comp.type},identifiers={','.join(map(str, comp.attributes.get('identifiers', [])))})"


def relation_pretty_short(rel: RelationWrapper):
    return f"#{rel.source}->[type={rel.type}]->{rel.target}"


def components_short_print(comps, delimiter="\n\t\t"):
    return delimiter.join(map(component_pretty_short, comps))


def relations_short_print(rels, delimiter="\n\t\t"):
    return delimiter.join(map(relation_pretty_short, rels))
