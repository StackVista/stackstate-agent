#!/usr/bin/env python3
"""
Specialized resolver for pkg/collector/corechecks/system/disk/diskv2/disk{,_nix}_test.go.

Pattern variants in these two files:
  A) Direct `Configure(...)` call. HEAD changed style (uses senderManager local var
     and adds trailing "provider"/"" arg). STS adds handler.NewMockCheckManager()
     at position 2 and wraps data in integration.Data{}.
     Resolution: take HEAD verbatim, insert `handler.NewMockCheckManager(), `
     after the first `,` inside the call.

  B) HEAD replaced inline setup with helper `configureCheck(t, diskCheck, ...)`.
     Resolution: take HEAD as-is. (configureCheck helper itself is patched
     separately to inject handler.NewMockCheckManager().)

  Other: leave conflict markers in place for hand resolution.
"""
import re
import sys
from pathlib import Path

CONFLICT_RE = re.compile(
    r'<<<<<<< HEAD\n(.*?)\n\|\|\|\|\|\|\| [^\n]*\n(.*?)\n=======\n(.*?)\n>>>>>>> [^\n]+\n',
    re.DOTALL,
)


def insert_check_manager(line):
    """Insert `handler.NewMockCheckManager(), ` after first `,` in any `Configure(` call on the line."""
    # Find positions of `.Configure(` in the line. (could be `diskCheck.Configure(` etc.)
    out = []
    i = 0
    while i < len(line):
        m = re.search(r'\bConfigure\(', line[i:])
        if not m:
            out.append(line[i:])
            break
        start = i + m.start()
        open_paren = i + m.end()  # position right after `(`
        # If the next non-space token is `senderManager` (HEAD style) we're at a call site.
        # Find the first `,` inside this Configure call (at depth 1).
        depth = 1
        j = open_paren
        first_comma = -1
        while j < len(line) and depth > 0:
            c = line[j]
            if c == '(':
                depth += 1
            elif c == ')':
                depth -= 1
                if depth == 0:
                    break
            elif c == ',' and depth == 1 and first_comma == -1:
                first_comma = j
                break
            j += 1
        if first_comma == -1:
            out.append(line[i:open_paren])
            i = open_paren
            continue
        # Insert after the comma + space
        out.append(line[i:first_comma + 1])
        out.append(' handler.NewMockCheckManager(),')
        i = first_comma + 1
    return ''.join(out)


def try_resolve_pattern_a(head):
    """If every non-blank head line contains Configure(, insert checkManager into each. Return new block or None."""
    lines = head.split('\n')
    has_any = False
    new_lines = []
    for ln in lines:
        if 'Configure(' in ln and 'configureCheck(' not in ln:
            has_any = True
            new_lines.append(insert_check_manager(ln))
        else:
            new_lines.append(ln)
    if not has_any:
        return None
    return '\n'.join(new_lines) + '\n'


def try_resolve_pattern_b(head):
    """If HEAD uses configureCheck(...) helper, take HEAD as-is."""
    if 'configureCheck(' in head:
        return head + '\n'
    return None


def process_file(path):
    text = Path(path).read_text()
    resolved = 0
    skipped = 0
    skipped_samples = []

    def repl(m):
        nonlocal resolved, skipped
        head, _anc, sts = m.group(1), m.group(2), m.group(3)
        # Try pattern A first (more specific)
        r = try_resolve_pattern_a(head)
        if r is None:
            r = try_resolve_pattern_b(head)
        if r is None:
            skipped += 1
            if len(skipped_samples) < 3:
                skipped_samples.append(head[:120].replace('\n', ' | '))
            return m.group(0)
        resolved += 1
        return r

    new_text = CONFLICT_RE.sub(repl, text)
    if new_text != text:
        Path(path).write_text(new_text)
    return resolved, skipped, skipped_samples


def main():
    for path in sys.argv[1:]:
        r, s, samples = process_file(path)
        print(f"{path}: resolved={r} skipped={s}")
        for x in samples:
            print(f"   skipped sample: {x}")


if __name__ == "__main__":
    main()
