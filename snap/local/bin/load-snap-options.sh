#!/bin/bash -e

export ARGS=$(snapctl get args)

exec "$@"
