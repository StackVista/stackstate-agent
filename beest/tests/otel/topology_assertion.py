import re


def component_pretty_short(comp: dict):
    return f"#{comp['id']}#[{comp['name']}](type={comp['type']},identifiers={','.join(map(str, comp['identifiers']))})"


def relation_pretty_short(rel: dict):
    return f"#{rel['source']}->[type={rel['type']}]->{rel['target']}"


def components_short_print(comps, delimiter="\n\t\t"):
    return delimiter.join(map(component_pretty_short, comps))


def relations_short_print(rels, delimiter="\n\t\t"):
    return delimiter.join(map(relation_pretty_short, rels))


class TopologyMatcherResult:
    def __init__(self, components, relations):
        self.components = components
        self.relations = relations


class TopologyMatcher:
    def __init__(self):
        self.components: list[ComponentMatcher] = []
        self.relations: list[RelationMatcher] = []

    def component(self, id: str, **kwargs) -> 'TopologyMatcher':
        self.components.append(ComponentMatcher(id, kwargs))
        return self

    def one_way_direction(self, source: str, target: str, **kwargs) -> 'TopologyMatcher':
        kwargs['dependencyDirection'] = 'ONE_WAY'
        self.relations.append(RelationMatcher(source, target, kwargs))
        return self

    def find(self, topology_result):
        # find all matching components and group them by virtual node (id) from a pattern
        matching_components = {}
        for comp_match in self.components:
            found = False
            for component in topology_result['components']:
                if comp_match.match(component):
                    if comp_match.id not in matching_components:
                        matching_components[comp_match.id] = [component]
                    else:
                        matching_components[comp_match.id].append(component)
                    found = True
            assert found, f"component {comp_match} has not been not found"

        relation_by_id = {}
        for relation in topology_result['relations']:
            relation_by_id[relation['id']] = relation

        for comp_rel in self.relations:
            source_candidates = matching_components.get(comp_rel.source, [])
            target_candidates = matching_components.get(comp_rel.target, [])
            assert len(source_candidates) > 0 and len(target_candidates) > 0, \
                f"relation {comp_rel} has not been found,\n" \
                f"\tsource candidates:\n\t\t{components_short_print(source_candidates)}\n" \
                f"\ttarget candidates:\n\t\t{components_short_print(target_candidates)}\n"

            source_relations = set([id for source in source_candidates for id in source.get('outgoingRelations', [])])
            target_relations = set([id for target in target_candidates for id in target.get('incomingRelations', [])])
            relation_candidate_ids = list(source_relations & target_relations)
            relation_candidates = [relation_by_id[id] for id in relation_candidate_ids if id in relation_by_id]
            found = False
            for relation in relation_candidates:
                if comp_rel.match(relation):
                    found = True

            assert found, \
                f"relation {comp_rel} has not been matched,\n" \
                f"\tsource candidates:\n\t\t{components_short_print(source_candidates)}\n" \
                f"\ttarget candidates:\n\t\t{components_short_print(target_candidates)}\n" \
                f"\tcandidate relations:\n\t\t{relations_short_print(relation_candidates)}"
        # ensure isomorphism with a assertion graph

        result_components = {}
        for k, v in matching_components.items():
            result_components[k] = v[0]

        return TopologyMatcherResult(result_components, [])


class StringPropertyMatcher:
    def __init__(self, key, value):
        self.key = key
        self.value = value

    def __str__(self):
        return f"{self.key}~={self.value}"

    def match(self, value):
        return re.fullmatch(self.value, value[self.key])


class ComponentMatcher:
    def __init__(self, id: str, props: dict):
        self.id = id
        self.matchers = []
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def __str__(self):
        return f"{self.id}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, component: dict):
        for m in self.matchers:
            if not m.match(component):
                return False
        return True


class RelationMatcher:
    def __init__(self, source: str, target: str, props: dict):
        self.source = source
        self.target = target
        self.matchers = []
        for k, v in props.items():
            self.matchers.append(StringPropertyMatcher(k, v))

    def __str__(self):
        return f"{self.source}->{self.target}[{','.join([str(m) for m in self.matchers])}]"

    def match(self, relation: dict):
        for m in self.matchers:
            if not m.match(relation):
                return False
        return True
