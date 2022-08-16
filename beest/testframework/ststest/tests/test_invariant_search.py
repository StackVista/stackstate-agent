from unittest import TestCase
from .context import ConsistentGraphMatcher, no_conflict


class TestInvariantSearch(TestCase):
    def test_no_conflict(self):
        self.assertTrue(no_conflict({'a': 'a1'}, {'b': 'b1'}))
        self.assertTrue(no_conflict({'a': 'a1', 'b': 'b1'}, {'b': 'b1'}))
        self.assertTrue(no_conflict({'a': 'a1'}, {'a': 'a1', 'b': 'b1'}))
        self.assertFalse(no_conflict({'a': 'a1'}, {'a': 'a2'}))
        self.assertFalse(no_conflict({'a': 'a1', 'b': 'b1'}, {'b': 'b2'}))
        self.assertFalse(no_conflict({'a': 'a1'}, {'a': 'a2', 'b': 'b2'}))

    def test_tuple_as_key(self):
        invm = ConsistentGraphMatcher()
        invm.add_choice_of_spec([{'a': 'a1'}, {'a': 'a2'}])
        #                                    \ - valid path
        invm.add_choice_of_spec([{'b': 'b1'}, {'b': 'b2'}])
        #                                    /
        invm.add_choice_of_spec([{'c': 'c1'}, {'c': 'c2'}])
        invm.add_choice_of_spec([{'a': 'a1', 'b': 'b1', ('a', 'b'): 'a1b1'}, {'a': 'a1', 'b': 'b2', ('a', 'b'): 'a1b2'}])
        invm.add_choice_of_spec([{'b': 'b2', 'c': 'c1', ('b', 'c'): 'b2c1'}])
        result = invm.get_graphs()
        expected = [
            {'a': 'a1', 'b': 'b2', 'c': 'c1', ('a', 'b'): 'a1b2', ('b', 'c'): 'b2c1'},
        ]
        self.assertEqual(expected, result)

    def test_new_set_of_specs_removes_option(self):
        invm = ConsistentGraphMatcher()
        invm.add_choice_of_spec([{'a': 'a1', 'b': 'b1'}, {'a': 'a2', 'b': 'b1'}, {'a': 'a2', 'b': 'b2'}])
        invm.add_choice_of_spec([{'b': 'b2', 'c': 'c1'}, {'b': 'b1', 'c': 'c2'}])
        result1 = invm.get_graphs()
        expected1 = [
            {'a': 'a2', 'b': 'b2', 'c': 'c1'},
            {'a': 'a1', 'b': 'b1', 'c': 'c2'},
            {'a': 'a2', 'b': 'b1', 'c': 'c2'},
        ]
        self.assertEqual(expected1, result1)

        invm.add_choice_of_spec([{'c': 'c2', 'a': 'a1'}, {'c': 'c1', 'a': 'a2'}])
        result2 = invm.get_graphs()
        expected2 = [
            {'a': 'a1', 'b': 'b1', 'c': 'c2'},
            {'a': 'a2', 'b': 'b2', 'c': 'c1'},
        ]
        self.assertEqual(expected2, result2)

    def test_ambiguity(self):
        invm = ConsistentGraphMatcher()
        invm.add_choice_of_spec([{'p1': 'pod-a'}, {'p1': 'pod-b'}])
        invm.add_choice_of_spec([{'p2': 'pod-a'}, {'p2': 'pod-b'}])
        result1 = invm.get_graphs()
        expected1 = [
            {'p1': 'pod-b', 'p2': 'pod-a'},
            {'p1': 'pod-a', 'p2': 'pod-b'},
            # other combinations of specs are not valid
            # {'p1': 'pod-a', 'p2': 'pod-a'},
            # {'p1': 'pod-b', 'p2': 'pod-b'},
        ]
        self.assertEqual(expected1, result1)


