#!/usr/bin/env python3
"""Fail when a CVE exception under exceptions/ has passed its expires date.

The image scan already treats an expired exception as a live finding, but it runs
in `inform` mode here, so its exit code is always 0 and the dates never block
anything. The gating chart scan in StackVista/cve-reporter does not read this
tree at all. This check is what makes `expires` a real deadline.
"""

import argparse
import collections.abc
import datetime
import pathlib
import re
import sys

import yaml

DEFAULT_WARN_DAYS = 14
EXPIRES_PATTERN = re.compile(r"\d{4}-\d{2}-\d{2}\Z")


def load(path):
    with path.open(encoding="utf-8") as handle:
        return yaml.safe_load(handle) or {}


def check(path, today, warn_days):
    """Return (errors, warnings) for one exception file."""
    try:
        doc = load(path)
    except (yaml.YAMLError, OSError) as exc:
        return [f"{path}: cannot be parsed: {exc}"], []

    if not isinstance(doc, collections.abc.Mapping):
        return [f"{path}: document must be a YAML mapping"], []

    vulnerability = doc.get("vulnerability") or {}
    if not isinstance(vulnerability, collections.abc.Mapping):
        return [f"{path}: vulnerability must be a YAML mapping"], []

    cve = vulnerability.get("id") or "<no vulnerability.id>"
    raw = doc.get("expires")

    if raw is None or str(raw).strip() == "":
        return [f"{path}: {cve} has no expires date"], []

    # The scan treats an unparseable date as already expired, so an exception
    # that looks valid but is not parseable must fail here too rather than
    # sitting in the tree looking effective.
    if type(raw) is datetime.date:
        expires = raw
    elif isinstance(raw, str) and EXPIRES_PATTERN.fullmatch(raw):
        try:
            expires = datetime.date.fromisoformat(raw)
        except ValueError:
            return [f"{path}: {cve} has an unparseable expires date {raw!r} (want YYYY-MM-DD)"], []
    else:
        return [f"{path}: {cve} has an unparseable expires date {raw!r} (want YYYY-MM-DD)"], []

    # Same boundary as the scan evaluator: valid through the expires date, dead
    # the day after.
    if today > expires:
        days = (today - expires).days
        return [
            f"{path}: {cve} expired {days} day(s) ago on {expires.isoformat()} — "
            "re-verify upstream and either renew with a new date or drop the exception"
        ], []

    remaining = (expires - today).days
    if remaining <= warn_days:
        return [], [f"{path}: {cve} expires in {remaining} day(s) on {expires.isoformat()}"]

    return [], []


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--exceptions-dir",
        default=str(pathlib.Path(__file__).resolve().parent.parent / "exceptions"),
        help="Directory tree of exception YAML files.",
    )
    parser.add_argument(
        "--warn-days",
        type=int,
        default=DEFAULT_WARN_DAYS,
        help=f"Warn when an exception expires within this many days (default {DEFAULT_WARN_DAYS}).",
    )
    args = parser.parse_args()

    root = pathlib.Path(args.exceptions_dir)
    if not root.is_dir():
        print(f"ERROR: exceptions directory {root} does not exist")
        return 1

    # The scan evaluator resolves expiry against UTC; matching it keeps the two
    # from disagreeing for a few hours a day.
    today = datetime.datetime.now(datetime.timezone.utc).date()

    errors = []
    warnings = []
    paths = sorted(root.rglob("*.yaml"))
    for path in paths:
        file_errors, file_warnings = check(path, today, args.warn_days)
        errors.extend(file_errors)
        warnings.extend(file_warnings)

    for warning in warnings:
        print(f"WARNING: {warning}")
    for error in errors:
        print(f"ERROR: {error}")

    print(f"Checked {len(paths)} exception file(s) against {today.isoformat()}.")
    if errors:
        print(f"{len(errors)} exception(s) are expired or malformed.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
