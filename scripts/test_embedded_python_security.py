"""Check the StringPrep backport using the packaged Python interpreter."""

import hashlib
import stringprep
import sys
import unittest
from pathlib import Path


class EmbeddedPythonSecurityTests(unittest.TestCase):
    def test_idna_uses_unicode_3_2_case_folding(self):
        cases = (
            ("\N{CHEROKEE LETTER A}\N{CHEROKEE LETTER A}", b"xn--58da"),
            ("\N{GEORGIAN CAPITAL LETTER AN}.", b"xn--7md."),
            ("\N{CYRILLIC LETTER PALOCHKA}.example", b"xn--d5a.example"),
            ("\N{ROMAN NUMERAL REVERSED ONE HUNDRED}.example.", b"xn--q5g.example."),
        )
        for name, encoded in cases:
            with self.subTest(name=name):
                self.assertEqual(name.encode("idna"), encoded)
        self.assertEqual("example.invalid".encode("idna"), b"example.invalid")


if __name__ == "__main__":
    print(sys.version)
    module = Path(stringprep.__file__)
    print(f"{module}: {hashlib.sha256(module.read_bytes()).hexdigest()}")
    unittest.main()
