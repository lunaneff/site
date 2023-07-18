---
title: Projects
type: page
---

Here's some things I made:

## This website
Made with a custom static site generator, written in 236 lines of Go. Definitely not as flexible as something like
Hugo, but a lot simpler.

[Git repo](https://git.lunaa.ch/luna/site) - files in the root directory are build infrastructure and directories
are content

## Nix configs
These are my Nix configs, defining some of my devices. Not intended to be used by others, but you can look at it
for inspiration for your config.

Probably the most interesting to others would be my [Nix cache's configuration](https://git.lunaa.ch/luna/nix-configs/src/branch/main/hosts/nix-cache/configuration.nix),
as it contains a build script to build everything there is in a set of repos (auto-build via webhook TBD)

[Git repo](https://git.lunaa.ch/luna/nix-configs)

## Windows fonts for Nix
A set of Nix packages for the fonts bundled in Windows 10 and 11. These packages work in a very unique way: during
the build process, they boot up a lightweight VM, "mount" the URL to the install ISO via HTTPS, and then mount the
ISO using a loop device to finally extract the fonts from install.wim into the host's Nix store.

[Git repo](https://git.lunaa.ch/luna/nix-windows-fonts)
