#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "== Building C library =="
gcc -shared -fPIC -O2 -o calculator_server/libcalculator.so c_lib/calculator.c
echo "  -> calculator_server/libcalculator.so"

echo "== Building Rust library =="
(cd rust_lib && cargo build --release)
cp rust_lib/target/release/libcalculator_rust.so calculator_server/libcalculator_rust.so
echo "  -> calculator_server/libcalculator_rust.so"

echo "Build complete."
