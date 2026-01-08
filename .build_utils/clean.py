#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3

"""
Cleans all subdirectories with no errors, language agnostic.
"""

import shutil
from os import listdir
from os.path import isdir
from pathlib import Path


def safe_rmtree(path):
    """Safely remove directory tree."""
    try:
        if Path(path).exists():
            shutil.rmtree(path)
            # print(f"✓ Removed {path}")
        else:
            pass
            # print(f"Directory {path} doesn't exist")
    except PermissionError:
        print(f"✗ Permission denied: {path}")
    except Exception as e:
        print(f"✗ Error removing {path}: {e}")


if __name__ == "__main__":
    for dir in listdir():
        if isdir(dir):
            safe_rmtree(dir + "/target")
            safe_rmtree(dir + "/debug")
            safe_rmtree(dir + "/.venv")
            safe_rmtree(dir + "/dist")
