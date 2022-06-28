from unittest import TestCase
from .context import InvariantSearch, no_conflict


class TestInvariantSearch(TestCase):
    def test_no_conflict(self):
        self.assertTrue(no_conflict({'a': 'a1'}, {'b': 'b1'}))
        self.assertTrue(no_conflict({'a': 'a1', 'b': 'b1'}, {'b': 'b1'}))
        self.assertTrue(no_conflict({'a': 'a1'}, {'a': 'a1', 'b': 'b1'}))
        self.assertFalse(no_conflict({'a': 'a1'}, {'a': 'a2'}))
        self.assertFalse(no_conflict({'a': 'a1', 'b': 'b1'}, {'b': 'b2'}))
        self.assertFalse(no_conflict({'a': 'a1'}, {'a': 'a2', 'b': 'b2'}))

    def test_one(self):
        invm = InvariantSearch()
        invm \
            .add_partial_invariant({'a': 'a1', 'b': 'b1'}) \
            .add_partial_invariant({'a': 'a2', 'b': 'b1'}) \
            .add_partial_invariant({'a': 'a2', 'b': 'b2'}) \
            .add_partial_invariant({'b': 'b2', 'c': 'c1'}) \
            .add_partial_invariant({'b': 'b1', 'c': 'c2'}) \
            .add_partial_invariant({'c': 'c2', 'a': 'a1'})

        invariants = invm.get_invariants(partial=True)
        expected = [
            {'a': 'a1', 'b': 'b1'},

            {'a': 'a2', 'b': 'b1'},

            {'a': 'a2', 'b': 'b2'},

            {'b': 'b2', 'c': 'c1'},
            {'a': 'a2', 'b': 'b2', 'c': 'c1'},

            {'b': 'b1', 'c': 'c2'},
            {'a': 'a1', 'b': 'b1', 'c': 'c2'},
            {'a': 'a2', 'b': 'b1', 'c': 'c2'},

            {'a': 'a1', 'c': 'c2'},
        ]
        self.assertEqual(expected, invariants)

        expected = [
            {'a': 'a2', 'b': 'b2', 'c': 'c1'},
            {'a': 'a1', 'b': 'b1', 'c': 'c2'},
            {'a': 'a2', 'b': 'b1', 'c': 'c2'},
        ]
        total_invariants = invm.get_invariants()
        self.assertEqual(expected, total_invariants)



