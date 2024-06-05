#!/usr/bin/env bash

declare -a __sources=(common ui mesh)

for source in "${__sources[@]}"; do
  source "$(dirname "${BASH_SOURCE[0]}")/${source}.bash"
done

unset __sources
