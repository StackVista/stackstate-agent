#!/usr/bin/env python3
"""
Batch resolver for the "Configure-pattern" merge conflicts in the 7.71.2 -> 7.78.2 upstream merge.

The pattern: DD added a parameter (last position) and STS added a parameter (position 2-ish) to the
same Configure() function. Both signature definitions and call sites are affected.

Algorithm (per single-line conflict):
  1. Compute the longest common SUFFIX of HEAD, ancestor, and STS lines.
  2. Strip that suffix from all three. Compute longest common PREFIX of the three middles.
  3. Strip prefix. We now have head_mid, anc_mid, sts_mid.
  4. If anc_mid is a SUFFIX of head_mid (HEAD added something purely at the end of the middle):
       head_addition = head_mid[: len(head_mid) - len(anc_mid)]
       resolution    = prefix + sts_mid + head_addition + suffix
  5. Otherwise the pattern doesn't apply — leave the conflict markers in place.

This handles both:
  - Definitions: "func ...(senderManager, source string) error {" -> add ", provider string"
  - Call sites:  "obj.Configure(sender, ..., \"test\")" -> add ", \"provider\""

Multi-line conflict blocks are not touched (left for manual resolution).

Usage: ./resolve_configure_pattern.py <files...>
Reports per file: resolved hunks, skipped hunks (multi-line or pattern mismatch).
"""
import re
import sys
from pathlib import Path

CONFLICT_RE = re.compile(
    r'<<<<<<< [^\n]*\n'
    r'(.*?)'                           # group 1: head/ours section
    r'\|\|\|\|\|\|\| [^\n]*\n'
    r'(.*?)'                           # group 2: ancestor/base section
    r'=======\n'
    r'(.*?)'                           # group 3: sts/theirs section
    r'>>>>>>> [^\n]*\n',
    re.DOTALL,
)


def common_prefix(strings):
    if not strings:
        return ""
    s_min = min(strings, key=len)
    for i, c in enumerate(s_min):
        for s in strings:
            if s[i] != c:
                return s_min[:i]
    return s_min


def common_suffix(strings):
    if not strings:
        return ""
    s_min = min(strings, key=len)
    n = len(s_min)
    for i in range(1, n + 1):
        c = s_min[-i]
        for s in strings:
            if s[-i] != c:
                return s_min[-(i - 1):] if i > 1 else ""
    return s_min


def resolve_single_line(head_l, anc_l, sts_l):
    """Resolve one line of a conflict where DD added text near the end of an existing line."""
    cs = common_suffix([head_l, anc_l])
    head_stripped = head_l[: len(head_l) - len(cs)]
    anc_stripped = anc_l[: len(anc_l) - len(cs)]

    if not head_stripped.startswith(anc_stripped):
        return None
    head_addition = head_stripped[len(anc_stripped):]
    if not head_addition:
        # No DD addition on this line — line is unchanged in HEAD vs ancestor.
        # Take STS line as-is (STS's modification stands).
        return sts_l

    if not sts_l.endswith(cs):
        return None
    sts_pre = sts_l[: len(sts_l) - len(cs)]
    return sts_pre + head_addition + cs


def looks_like_import_block(lines):
    """True iff every non-blank line looks like a Go import statement (with leading whitespace)."""
    has_import = False
    for line in lines:
        s = line.strip()
        if not s:
            continue
        # Go imports: optional alias, then quoted path. e.g. `"foo/bar"` or `alias "foo/bar"` or `_ "foo"`
        if not (s.endswith('"') and s.count('"') >= 2):
            return False
        has_import = True
    return has_import


def try_resolve_import_block(head, ancestor, sts):
    """Set-union resolution: HEAD ∪ (STS - ancestor). Drops imports DD removed; keeps STS's adds."""
    def split_lines(s):
        ls = s.split('\n')
        if ls and ls[-1] == '':
            ls = ls[:-1]
        return ls

    head_lines = split_lines(head)
    anc_lines = split_lines(ancestor)
    sts_lines = split_lines(sts)

    # All three (or just the non-empty ones) must look like import blocks.
    non_empty = [ls for ls in [head_lines, anc_lines, sts_lines] if ls]
    if not non_empty:
        return None
    if not all(looks_like_import_block(ls) for ls in non_empty):
        return None

    # STS's additions vs ancestor (preserving STS's order, dedup against HEAD).
    anc_set = {l.strip() for l in anc_lines if l.strip()}
    head_set = {l.strip() for l in head_lines if l.strip()}
    sts_additions = []
    for line in sts_lines:
        s = line.strip()
        if s and s not in anc_set and s not in head_set:
            sts_additions.append(line)

    result_lines = list(head_lines) + sts_additions
    if not result_lines:
        return None
    return '\n'.join(result_lines) + '\n'


def try_resolve(head, ancestor, sts):
    # First try the import-block union resolver (it's a different shape than line-aligned).
    r = try_resolve_import_block(head, ancestor, sts)
    if r is not None:
        return r

    # Otherwise: line-aligned Configure-pattern resolver.
    head_lines = head.split('\n')
    anc_lines = ancestor.split('\n')
    sts_lines = sts.split('\n')

    if head_lines and head_lines[-1] == '':
        head_lines = head_lines[:-1]
    if anc_lines and anc_lines[-1] == '':
        anc_lines = anc_lines[:-1]
    if sts_lines and sts_lines[-1] == '':
        sts_lines = sts_lines[:-1]

    if not (len(head_lines) == len(anc_lines) == len(sts_lines)):
        return None
    if len(head_lines) == 0:
        return None

    resolved_lines = []
    for h, a, s in zip(head_lines, anc_lines, sts_lines):
        r = resolve_single_line(h, a, s)
        if r is None:
            return None
        resolved_lines.append(r)
    return '\n'.join(resolved_lines) + '\n'


def process_file(path):
    text = Path(path).read_text()
    resolved = 0
    skipped = 0

    def repl(m):
        nonlocal resolved, skipped
        head, ancestor, sts = m.group(1), m.group(2), m.group(3)
        result = try_resolve(head, ancestor, sts)
        if result is None:
            skipped += 1
            return m.group(0)
        resolved += 1
        return result

    new_text = CONFLICT_RE.sub(repl, text)
    if new_text != text:
        Path(path).write_text(new_text)
    return resolved, skipped


def main():
    total_files = 0
    total_resolved = 0
    total_skipped = 0
    fully_resolved_files = []
    partially_resolved_files = []
    untouched_files = []
    for path in sys.argv[1:]:
        resolved, skipped = process_file(path)
        total_files += 1
        total_resolved += resolved
        total_skipped += skipped
        if resolved and not skipped:
            fully_resolved_files.append(path)
        elif resolved and skipped:
            partially_resolved_files.append((path, resolved, skipped))
        else:
            untouched_files.append(path)

    print(f"Files processed: {total_files}")
    print(f"Hunks resolved:  {total_resolved}")
    print(f"Hunks skipped:   {total_skipped}")
    print()
    print(f"Fully resolved files: {len(fully_resolved_files)}")
    print(f"Partially resolved (still have conflicts): {len(partially_resolved_files)}")
    for p, r, s in partially_resolved_files[:20]:
        print(f"  {p}  ({r} resolved, {s} skipped)")
    print(f"Untouched files (no Configure pattern matched): {len(untouched_files)}")
    for p in untouched_files[:20]:
        print(f"  {p}")


if __name__ == "__main__":
    main()
