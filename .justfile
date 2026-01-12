alias r := run
alias b := build
alias c := clean
alias d := develop
alias n := new

# default target, lists all targets available
list:
    @just --list

# runs the default master nix flake target
run:
    @nix run

# builds the nix targets
build:
    @nix build

# removes all build artifacts and binaries
clean:
    @rm -rf result target
    @ ./.build_utils/clean.py

# runs clean and destroys the machine's nix cache (used to free all system storage memory related to this project)
nuke: clean
    @nix-collect-garbage

# creates a new node from the script, with NAME as the node name, and BUILD as the build type (currently python or rust)
new NAME BUILD:
    @ ./.build_utils/new_node.py {{ NAME }} {{ BUILD }}
    @nix flake lock

# opens the nix shell
develop:
    @nix develop
