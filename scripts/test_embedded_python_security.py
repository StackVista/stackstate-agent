"""Regressions to run with the Python interpreter shipped in the agent image."""

import hashlib
import io
import poplib
import stringprep
import sys
import tarfile
import tempfile
import unittest
import urllib.request
import zipfile
from pathlib import Path
from unittest.mock import Mock, patch


class EmbeddedPythonSecurityTests(unittest.TestCase):
    def test_credentials_are_scoped_by_scheme(self):
        for manager in (
            urllib.request.HTTPPasswordMgr,
            urllib.request.HTTPPasswordMgrWithDefaultRealm,
            urllib.request.HTTPPasswordMgrWithPriorAuth,
        ):
            with self.subTest(manager=manager.__name__):
                passwords = manager()
                passwords.add_password(None, "https://example.invalid/", "user", "test-value")
                self.assertEqual(
                    passwords.find_user_password(None, "https://example.invalid/"),
                    ("user", "test-value"),
                )
                self.assertEqual(passwords.find_user_password(None, "http://example.invalid/"), (None, None))
                passwords.add_password(None, "proxy.invalid", "proxy", "test-value")
                for scheme in ("http", "https"):
                    self.assertEqual(
                        passwords.find_user_password(None, f"{scheme}://proxy.invalid/"),
                        ("proxy", "test-value"),
                    )

    def test_tar_filters_do_not_create_directories_outside_destination(self):
        for extraction_filter in ("tar", "data"):
            with self.subTest(filter=extraction_filter), tempfile.TemporaryDirectory() as root:
                destination = Path(root) / "destination"
                destination.mkdir()
                archive = io.BytesIO()
                content = b"expected content"
                with tarfile.open(fileobj=archive, mode="w") as writer:
                    member = tarfile.TarInfo("../outside/../destination/sub/file")
                    member.size = len(content)
                    writer.addfile(member, io.BytesIO(content))
                archive.seek(0)
                with tarfile.open(fileobj=archive) as reader:
                    reader.extractall(destination, filter=extraction_filter)
                self.assertEqual((destination / "sub/file").read_bytes(), content)
                self.assertFalse((Path(root) / "outside").exists())

    def test_zipfile_small_reads_bound_decompression(self):
        content = b"\0" * (4 * 1024 * 1024)
        for compression in (zipfile.ZIP_BZIP2, zipfile.ZIP_LZMA):
            with self.subTest(compression=compression):
                archive = io.BytesIO()
                with zipfile.ZipFile(archive, "w", compression=compression) as writer:
                    writer.writestr("content", content)
                archive.seek(0)
                with zipfile.ZipFile(archive) as reader, reader.open("content") as member:
                    first = member._read1(100)
                    self.assertLessEqual(len(first), member.MIN_READ_SIZE)
                    self.assertEqual(first + member.read(), content)

    def test_tar_link_fallback_honors_filter_rejection(self):
        with tempfile.TemporaryDirectory() as destination:
            archive = io.BytesIO()
            with tarfile.open(fileobj=archive, mode="w") as writer:
                symlink = tarfile.TarInfo("a/b/s")
                symlink.type = tarfile.SYMTYPE
                symlink.linkname = "../escape"
                writer.addfile(symlink)
                hardlink = tarfile.TarInfo("q")
                hardlink.type = tarfile.LNKTYPE
                hardlink.linkname = "a/b/s"
                writer.addfile(hardlink)
            rejected = []

            def skip_unsafe(member, path):
                try:
                    return tarfile.data_filter(member, path)
                except tarfile.FilterError:
                    rejected.append(member.name)
                    return None

            archive.seek(0)
            with (
                tarfile.open(fileobj=archive) as reader,
                patch("tarfile.os.link", side_effect=OSError("Exercise link fallback")),
            ):
                reader.extractall(destination, filter=skip_unsafe)
            self.assertIn("q", rejected)
            self.assertTrue((Path(destination) / "a/b/s").is_symlink())
            self.assertFalse((Path(destination) / "q").is_symlink())
            self.assertFalse((Path(destination) / "q").exists())

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

    def test_pop3_rejects_control_characters_before_sending(self):
        client = poplib.POP3.__new__(poplib.POP3)
        client._debugging = 0
        client.encoding = "utf-8"
        client._putline = Mock()
        client._putcmd("USER valid-user")
        client._putline.assert_called_once_with(b"USER valid-user")
        client._putline.reset_mock()
        for character in (*range(32), 127):
            with self.subTest(character=character), self.assertRaises(ValueError):
                client._putcmd(f"USER invalid{chr(character)}value")
        client._putline.assert_not_called()


if __name__ == "__main__":
    print(f"Embedded interpreter: {sys.executable}; version: {sys.version}", flush=True)
    for module in (urllib.request, tarfile, poplib, stringprep, zipfile):
        source = Path(module.__file__)
        print(f"{module.__name__}: {source}; sha256={hashlib.sha256(source.read_bytes()).hexdigest()}", flush=True)
    unittest.main(verbosity=2)
