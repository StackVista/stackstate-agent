
def no_conflict(d1: dict, d2: dict):
    if len(d2) < len(d1):
        return no_conflict(d2, d1)
    for k, v1 in d1.items():
        if d2.get(k, v1) != v1:
            return False
    return True


class InvariantSearch:
    def __init__(self):
        self.invariants: list[dict] = []
        self.involved = set()

    def _key(k, v):
        return f"{k}={v}"

    def add_partial_invariant(self, partial_inv: dict):
        self.involved = self.involved.union(partial_inv.keys())
        extended = [partial_inv]
        for inv in self.invariants:
            if partial_inv != inv and no_conflict(partial_inv, inv):
                extended.append(inv | partial_inv)
        for inv in extended:
            if inv not in self.invariants:
                # TODO possible optimization: drop invariant that don't have all involved components
                self.invariants.append(inv)
        return self

    def get_invariants(self, partial: bool = False):
        if not partial:
            return [inv for inv in self.invariants if inv.keys() == self.involved]
        return self.invariants
