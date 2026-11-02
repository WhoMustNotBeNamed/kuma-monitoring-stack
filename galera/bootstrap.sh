#!/bin/bash

STATE=/var/lib/mysql/grastate.dat

if [ ! -f "$STATE" ] || grep -q '^safe_to_bootstrap: *1' "$STATE"; then
  set -- "$@" --wsrep-new-cluster
fi
exec docker-entrypoint.sh mariadbd "$@"
