#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3 git

import subprocess
import sys
from pathlib import Path

if len(sys.argv) != 3:
    print(f"Usage: {sys.argv[0]} <prev> <target>")
    sys.exit(1)

prev = sys.argv[1]
target = sys.argv[2]

cwd = Path.cwd()
prev_dir = cwd / prev
target_dir = cwd / target

# Get all tracked and untracked (but not ignored) files.
files = subprocess.check_output(
    ["git", "ls-files", "--cached", "--others", "--exclude-standard"],
    text=True,
).splitlines()

for rel_path in files:
    path = cwd / rel_path

    try:
        text = path.read_text()
    except (UnicodeDecodeError, OSError):
        # Skip binary files
        continue

    if prev in text:
        path.write_text(text.replace(prev, target))
        print(f"updated: {rel_path}")

# Rename files/directories whose names contain prev
for path in sorted(cwd.rglob("*"), key=lambda p: len(p.parts), reverse=True):
    if ".git" in path.parts:
        continue

    if prev in path.name:
        new_path = path.with_name(path.name.replace(prev, target))
        path.rename(new_path)
        print(f"renamed: {path} -> {new_path}")

# Rename top-level directory if requested
if prev_dir.exists():
    prev_dir.rename(target_dir)
    print(f"renamed: {prev_dir} -> {target_dir}")
