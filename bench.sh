#!/usr/bin/env bash

set -euo pipefail

hyperfine './bin/vai build' 'make'
