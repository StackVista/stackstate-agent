"""Check security backports using the packaged Python interpreter."""

import hashlib
import io
import platform
import stringprep
import sys
import tarfile
import tempfile
import unittest
from pathlib import Path

from build._compat.tarfile import safe_extractall


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

    def test_tarfile_hardlink_relocation(self):
        # CPython gh-157190's test_sneaky_hardlink_relocation, also exercised
        # through the packaged build frontend's extraction wrapper.
        for extraction in ("data", "tar", "build.safe_extractall"):
            with self.subTest(extraction=extraction), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                outside = root / "escape"
                outside.write_bytes(b"outside")
                before = outside.stat()
                destination = root / "dest"
                destination.mkdir()
                archive = io.BytesIO()
                with tarfile.open(fileobj=archive, mode="w") as tar:
                    member = tarfile.TarInfo("a/escape")
                    member.size = len(b"decoy")
                    tar.addfile(member, io.BytesIO(b"decoy"))
                    member = tarfile.TarInfo("a/b/s")
                    member.type = tarfile.SYMTYPE
                    member.linkname = "../escape"
                    tar.addfile(member)
                    member = tarfile.TarInfo("s")
                    member.type = tarfile.LNKTYPE
                    member.linkname = "a/b/s"
                    tar.addfile(member)
                archive.seek(0)
                with tarfile.open(fileobj=archive) as tar:
                    if extraction == "build.safe_extractall":
                        safe_extractall(tar, destination)
                    else:
                        tar.extractall(destination, filter=extraction)
                self.assertEqual((destination / "a/escape").read_bytes(), b"decoy")
                self.assertTrue((destination / "a/b/s").is_symlink())
                self.assertFalse((destination / "s").is_symlink())
                self.assertEqual((destination / "s").read_bytes(), b"decoy")
                self.assertEqual(outside.read_bytes(), b"outside")
                after = outside.stat()
                self.assertEqual((after.st_mode, after.st_mtime_ns), (before.st_mode, before.st_mtime_ns))


if __name__ == "__main__":
    print(sys.version)
    print(platform.machine())
    for imported in (stringprep, tarfile):
        module = Path(imported.__file__)
        print(f"{module}: {hashlib.sha256(module.read_bytes()).hexdigest()}")
    unittest.main(verbosity=2)
