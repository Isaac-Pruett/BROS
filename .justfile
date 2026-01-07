alias r := run
alias b := build
alias c := clean
alias d := develop

# default target, lists all targets available
list:
    @just --list

run:
    @nix run

build:
    @nix build

clean:
    @rm -rf result

nuke: clean
    @nix-collect-garbage

new-rust NAME:
    @mkdir ${{ NAME }}

new-py NAME:
    @mkdir ${{ NAME }}

develop:
    @nix develop
