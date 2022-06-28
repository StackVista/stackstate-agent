
def component_pretty_short(comp: dict):
    return f"#{comp['id']}#[{comp['name']}](type={comp['type']},identifiers={','.join(map(str, comp['identifiers']))})"


def relation_pretty_short(rel: dict):
    return f"#{rel['source']}->[type={rel['type']}]->{rel['target']}"


def components_short_print(comps, delimiter="\n\t\t"):
    return delimiter.join(map(component_pretty_short, comps))


def relations_short_print(rels, delimiter="\n\t\t"):
    return delimiter.join(map(relation_pretty_short, rels))
