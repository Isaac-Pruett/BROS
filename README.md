# BROS
---
In Development.

BROS is a framework for developing programs and processes that talk to each other using [Zenoh](https://zenoh.io), and have an automated build/launch process by using [Nix](https://nixos.org) flakes. It enables the user to use modern package managers ([cargo](https://doc.rust-lang.org/cargo/), [uv](https://docs.astral.sh/uv/)) with Zenoh's extremely fast Inter-Process-Communication. Developed initialally as a personal project with intent to use on robots and [PolyUAS](https://polyuas.org) autonomous drones.



## To use: 

Ensure that you have [Nix](https://nixos.org/download/) with [flake](https://nixos.wiki/wiki/Flakes) support installed on your machine.


### Running your first node:

```sh
nix develop
```
Then, after the shell has built:
```sh
simpledemo
```



### Running the Zenoh demo:
The next set of commands will run a rust and python node that will communicate over Zenoh.

```sh
nix run
```

or, alternatavely:

```sh
nix develop
```
Then, after the shell has built:
```sh
demo-ping-pong-zenoh
```


## Creating a new node
