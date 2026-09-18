"""Regressions to run with the Python interpreter shipped in the agent image."""

import hashlib
import io
import poplib
import sys
import tarfile
import tempfile
import unittest
import urllib.request
from pathlib import Path
from unittest.mock import Mock


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
    for module in (urllib.request, tarfile, poplib):
        source = Path(module.__file__)
        print(f"{module.__name__}: {source}; sha256={hashlib.sha256(source.read_bytes()).hexdigest()}", flush=True)
    unittest.main(verbosity=2)
