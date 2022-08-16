def no_conflict(d1: dict, d2: dict):
    if len(d2) < len(d1):
        return no_conflict(d2, d1)
    for k, v1 in d1.items():
        if d2.get(k, v1) != v1:
            return False
    return True


def no_double_match(d: dict):
    return len(d) == len(set(d.values()))


class ConsistentGraphMatcher:
    """
    ConsistentGraphMatcher computes a set of graphs
    where each one is consistent with one of the "node specifications" provided by `add_set_of_variants`

    Every invocation of `add_choice_of_spec` provides this class with a list of specifications
            (specification says that an abstract node A is a specific node A1)
    For example, you could say that node A is either A1 or A2:
    >>> cgm = ConsistentGraphMatcher()
    >>> cgm.add_choice_of_spec([{'A': 'A1'}, {'A': 'A2'}])
    ConsistentGraphMatcher will make sure that resulting graph will have either node A1 or node A2
    Next invocation `add_set_of_variants` tells ConsistentGraphMatcher that one of option should also be true in the end
    >>> cgm.add_choice_of_spec([{'B': 'B1'}, {'B': 'B2'}])
    >>> cgm.add_choice_of_spec([{'A': 'A1', 'B': 'B2'}, {'A': 'A2', 'B': 'B2'}])
    Note: that two relation to B2 are defined. So a graph with B=B1 is invalid.
    >>> cgm.get_graphs()
    """

    def __init__(self):
        self.specifications: list[list[dict]] = []

    def _key(k, v):
        return f"{k}={v}"

    def add_choice_of_spec(self, specs: list[dict]):
        """
        :param specs: list of specifications from virtual nodes A, B etc. to specific nodes A1, B1 etc.
                      a consistent graph contain one of the specifications
        :return:
        """
        self.specifications.append(specs)

    def get_graphs(self) -> list[dict]:
        """
        :return: list of specifications (e.g. {A: A1, B: B3, C: C1}) that have all previously specified virtual nodes
        """
        if len(self.specifications) == 0:
            return []

        specsets = self.specifications.copy()
        if len(specsets) == 0:
            return []

        valid_specs = specsets.pop()
        for specset in specsets:
            valid_specs = [
                combined
                for vspec in valid_specs
                for spec in specset
                if no_conflict(vspec, spec)
                   and no_double_match((combined := vspec | spec))
            ]

        return valid_specs
