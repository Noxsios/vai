#!/usr/bin/env bash

set -euo pipefail

hyperfine './bin/vai' 'make' -N --warmup 10

# hyperfine './bin/vai hello-world' 'make hello-world' -N --warmup 10
