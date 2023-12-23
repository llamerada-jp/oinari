#!/usr/bin/env bash

set -eux

if [ ! -e "build" ] ||  [ ! -e "node_modules" ]; then
  make setup
fi
make test
