#!/usr/bin/env python3
"""
Insert `handler.NewMockCheckManager(),` as 2nd arg in Configure(...) and CommonConfigure(...)
calls on the line. Used after upstream merges where DD adds a `handler.CheckManager` parameter
to the check.Check / CheckBase interfaces but doesn't update STS-specific test files.

Skips lines that already contain `NewMockCheckManager` (idempotent).
Adds the handler import if missing.

Usage: ./insert_check_manager.py <files...>
"""
import re
import sys
from pathlib import Path

CALL_RE = re.compile(r'\b(?:Configure|CommonConfigure)\(')

HANDLER_IMPORT = '"github.com/DataDog/datadog-agent/pkg/collector/check/handler"'


def insertions_for_line(line):
    """Return list of insertion offsets (positions to insert ` handler.NewMockCheckManager(),` AFTER)."""
    if 'NewMockCheckManager' in line:
        return []
    offsets = []
    for m in CALL_RE.finditer(line):
        # Find depth-1 first comma starting after `(`
        depth = 1
        i = m.end()
        while i < len(line) and depth > 0:
            c = line[i]
            if c == '(':
                depth += 1
            elif c == ')':
                depth -= 1
                if depth == 0:
                    break
            elif c == ',' and depth == 1:
                offsets.append(i + 1)  # insert AFTER the comma
                break
            i += 1
    return offsets


def patch_line(line):
    offs = insertions_for_line(line)
    if not offs:
        return line, 0
    # Insert from right to left so earlier offsets stay valid
    out = line
    for o in sorted(offs, reverse=True):
        out = out[:o] + ' handler.NewMockCheckManager(),' + out[o:]
    return out, len(offs)


def patch_file(path):
    text = Path(path).read_text()
    lines = text.splitlines(keepends=True)
    n = 0
    out = []
    for ln in lines:
        new, k = patch_line(ln)
        n += k
        out.append(new)
    new_text = ''.join(out)
    if new_text != text:
        if HANDLER_IMPORT not in new_text:
            # Add import to first import block
            m = re.search(r'(import \(\n)', new_text)
            if m:
                insert_pos = m.end()
                new_text = (new_text[:insert_pos] +
                            '\t' + HANDLER_IMPORT + ' // [sts] auto-inserted for NewMockCheckManager\n' +
                            new_text[insert_pos:])
        Path(path).write_text(new_text)
    return n


def main():
    total = 0
    for p in sys.argv[1:]:
        n = patch_file(p)
        print(f"{p}: {n} insertions")
        total += n
    print(f"Total: {total}")


if __name__ == '__main__':
    main()
