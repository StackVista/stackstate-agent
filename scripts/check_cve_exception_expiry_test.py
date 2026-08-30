import datetime
import pathlib
import tempfile
import unittest

import check_cve_exception_expiry


class CheckCveExceptionExpiryTest(unittest.TestCase):
    today = datetime.date(2026, 8, 30)

    def check_document(self, document, warn_days=14):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "exception.yaml"
            path.write_text(document, encoding="utf-8")
            return check_cve_exception_expiry.check(path, self.today, warn_days)

    def test_accepts_quoted_canonical_date(self):
        errors, warnings = self.check_document('expires: "2026-09-20"\n')
        self.assertEqual(errors, [])
        self.assertEqual(warnings, [])

    def test_accepts_unquoted_canonical_date(self):
        errors, warnings = self.check_document("expires: 2026-09-20\n")
        self.assertEqual(errors, [])
        self.assertEqual(warnings, [])

    def test_accepts_expiry_today(self):
        errors, warnings = self.check_document("expires: 2026-08-30\n")
        self.assertEqual(errors, [])
        self.assertEqual(len(warnings), 1)

    def test_rejects_expired_date(self):
        errors, warnings = self.check_document("expires: 2026-08-29\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_compact_date(self):
        errors, warnings = self.check_document("expires: 20260920\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_iso_week_date(self):
        errors, warnings = self.check_document("expires: 2026-W38-7\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_invalid_calendar_date(self):
        errors, warnings = self.check_document('expires: "2026-02-30"\n')
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_missing_date(self):
        errors, warnings = self.check_document("vulnerability:\n  id: CVE-2026-0001\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_malformed_yaml(self):
        errors, warnings = self.check_document("expires: [\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])

    def test_rejects_non_mapping_document(self):
        errors, warnings = self.check_document("- expires: 2026-09-20\n")
        self.assertEqual(len(errors), 1)
        self.assertEqual(warnings, [])


if __name__ == "__main__":
    unittest.main()
