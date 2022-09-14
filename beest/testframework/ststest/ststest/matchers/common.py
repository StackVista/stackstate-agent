from stscliv1 import ComponentWrapper


def get_common_relations(sources: list[ComponentWrapper], targets: list[ComponentWrapper]):
    # TODO consider BOTH_WAY type of relations
    source_relations = set([id for source in sources for id in source.outgoing_relations])
    target_relations = set([id for target in targets for id in target.incoming_relations])
    return list(source_relations & target_relations)


def filter_out_repeated_specs(specs: list[dict], ambiguous_elements: set[str]):
    def two_are_equivalent(spec1: dict, spec2: dict):
        for key in (set(spec1.keys()) | set(spec2.keys())):
            if key not in ambiguous_elements and spec1[key] != spec2[key]:
                return False
        return True

    distinct_specs = []
    for spec in specs:
        spec_is_distinct = True
        for dspec in distinct_specs:
            if two_are_equivalent(spec, dspec):
                spec_is_distinct = False
                break
        if spec_is_distinct:
            distinct_specs.append(spec)

    return distinct_specs

