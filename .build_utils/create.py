#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3

"""
Template duplicator and renamer
Creates a copy of TEMPLATE_NAME directory with new name,
renames directories/files and replaces content accordingly.
"""

import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path

TEMPLATES_DIRECTORY = ".build_utils/templates"


DEBUG = False


def dprint(text="\n"):
    if DEBUG:
        print(text)


def rename_content(content: str, old_name: str, new_name: str) -> str:
    """Replace all occurrences of old_name with new_name in text content"""
    # Case-sensitive exact match replacements
    content = content.replace(old_name, new_name)

    # Optional: also handle uppercase / capitalized versions
    content = content.replace(old_name.upper(), new_name.upper())
    content = content.replace(old_name.capitalize(), new_name.capitalize())

    return content


def copy_and_rename_tree(
    src_dir: Path, dst_dir: Path, old_name: str, new_name: str, blacklist: set[str]
):
    """
    Recursively copy directory tree while renaming paths and contents
    """
    # Create destination directory if it doesn't exist
    dst_dir.mkdir(parents=True, exist_ok=True)

    for item in src_dir.iterdir():
        # Skip blacklisted items (relative to template root)
        rel_path = item.relative_to(src_dir)
        if str(rel_path) in blacklist or rel_path.name in blacklist:
            dprint(f"  Skipping blacklisted: {rel_path}")
            continue

        new_name_part = item.name.replace(old_name, new_name)
        dst_path = dst_dir / new_name_part

        if item.is_dir():
            dprint(f"  Directory: {rel_path} → {new_name_part}")
            copy_and_rename_tree(item, dst_path, old_name, new_name, blacklist)
        else:
            # File - copy and rename content if text-like
            dprint(f"  File:      {rel_path} → {new_name_part}")
            shutil.copy2(item, dst_path)

            try:
                content = dst_path.read_text(encoding="utf-8")
                new_content = rename_content(content, old_name, new_name)
                if new_content != content:
                    dst_path.write_text(new_content, encoding="utf-8")
                    dprint(f"     ↳ content updated")
            except (UnicodeDecodeError, PermissionError):
                # Binary file or can't read → just keep the copy
                pass


def main():
    parser = argparse.ArgumentParser(
        description="Duplicate and rename a template directory"
    )
    parser.add_argument(
        "template_dir",
        type=str,
        help="Path to the template directory (folder named TEMPLATE_NAME)",
    )
    parser.add_argument("new_name", type=str, help="New project/module name")
    parser.add_argument(
        "--blacklist",
        nargs="*",
        default=[],
        help="Files/folders to skip (relative paths or names), e.g. .git node_modules venv",
    )

    args = parser.parse_args()

    src = (Path(TEMPLATES_DIRECTORY) / args.template_dir).resolve()
    old_name = src.name

    if not src.is_dir():
        print(f"Error: {src} is not a directory")
        return 1

    if old_name == args.new_name:
        print("Error: new name is the same as template name")
        return 1

    dst = Path(os.getcwd()) / args.new_name

    if dst.exists():
        print(f"Error: destination {dst} already exists")
        return 1

    print(f"Creating new project:")
    print(f"  From template : {src}")
    print(f"  New name      : {args.new_name}")
    print(f"  Destination   : {dst}")
    print()

    blacklist = set(args.blacklist)

    # Very common items you almost always want to skip
    default_blacklist = {
        ".git",
        "node_modules",
        "venv",
        ".venv",
        "__pycache__",
        "*.pyc",
        "dist",
        "build",
        ".pytest_cache",
        ".coverage",
        ".mypy_cache",
        ".ruff_cache",
        ".idea",
        ".vscode",
        ".DS_Store",
        "*.lock",
    }

    blacklist = blacklist | default_blacklist

    dprint("Blacklist:")
    for item in sorted(blacklist):
        dprint(f"  - {item}")
    dprint()

    copy_and_rename_tree(src, dst, old_name, args.new_name, blacklist)

    # ────────────────────────────────────────────────────────────────
    #  Run init.sh in the new project directory (if it exists)
    # ────────────────────────────────────────────────────────────────
    init_script = dst / "init.sh"

    if init_script.is_file():
        dprint(f"Found init.sh in new project → running it ...")
        print("Initializing...")
        print("─" * 60)

        try:
            # Run in the new directory
            result = subprocess.run(
                ["bash", "init.sh"],  # or ["./init.sh"] if you chmod +x it
                cwd=dst,  # ← important: working directory
                check=True,  # raise exception on non-zero exit
                text=True,
                capture_output=True,
            )

            print(result.stdout)
            if result.stderr:
                print("stderr:", result.stderr, file=sys.stderr)

            print("─" * 60)
            print("initialization finished successfully")

        except subprocess.CalledProcessError as e:
            print("┌──────────────────────────────────────────────────────┐")
            print("│                initialization failed                 │")
            print("└──────────────────────────────────────────────────────┘")
            print(f"Exit code: {e.returncode}")
            if e.stdout:
                print("stdout:\n", e.stdout)
            if e.stderr:
                print("stderr:\n", e.stderr, file=sys.stderr)
            print("\nNew project was still created, but initialization failed.")

        except FileNotFoundError:
            print("Warning: 'bash' not found — couldn't run init.sh")
        except Exception as e:
            print(f"Error while running init.sh: {e}")

        os.remove(Path(dst) / "init.sh")
    else:
        print("No init.sh found in new project — skipping initialization step")

    print("\nDone.")
    print(f"New project created at: {dst}")

    return 0


if __name__ == "__main__":
    exit(main())
