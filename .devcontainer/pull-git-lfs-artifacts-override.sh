#!/bin/sh

# Use this file to override /usr/local/share/pull-git-lfs-artifacts.sh to avoid conflicts with pre configured git hooks.
# Using this override disables some of the default git lfs behavior, such as automatically pulling LFS artifacts when cloning a repository. You may need to manually pull LFS artifacts using `git lfs pull` after cloning a repository.
