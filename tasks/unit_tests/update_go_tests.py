import unittest

from tasks.update_go import (
    PATTERN_MAJOR_MINOR,
    PATTERN_MAJOR_MINOR_BUGFIX,
    _extract_go_archive_checksums,
    _get_major_minor_version,
    _get_pattern,
)


class TestUpdateGo(unittest.TestCase):
    def test_get_minor_version(self):
        self.assertEqual(_get_major_minor_version("1.2.3"), "1.2")

    def test_get_pattern(self):
        self.assertEqual(_get_pattern("p+e", "p.st", is_bugfix=True), rf'(p\+e){PATTERN_MAJOR_MINOR_BUGFIX}(p\.st)')
        self.assertEqual(_get_pattern("p(re)", "p*st", is_bugfix=False), rf'(p\(re\)){PATTERN_MAJOR_MINOR}(p\*st)')

    def test_extract_go_archive_checksums(self):
        releases = [
            {
                "version": "go1.26.6",
                "files": [
                    {"os": "linux", "arch": "amd64", "sha256": "a" * 64},
                    {"os": "linux", "arch": "arm64", "sha256": "b" * 64},
                    {"os": "darwin", "arch": "arm64", "sha256": "c" * 64},
                ],
            }
        ]

        self.assertEqual(
            _extract_go_archive_checksums(releases, "1.26.6"),
            {"amd64": "a" * 64, "arm64": "b" * 64},
        )

    def test_extract_go_archive_checksums_requires_all_architectures(self):
        releases = [
            {
                "version": "go1.26.6",
                "files": [{"os": "linux", "arch": "amd64", "sha256": "a" * 64}],
            }
        ]

        with self.assertRaisesRegex(ValueError, "arm64"):
            _extract_go_archive_checksums(releases, "1.26.6")
