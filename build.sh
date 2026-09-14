#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "== Building C library =="
gcc -shared -fPIC -O2 -o bin/libcalculator.so c_lib/calculator.c
echo "  -> bin/libcalculator.so"

echo "== Building Rust library =="
(cd rust_lib && cargo build --release)
cp rust_lib/target/release/libcalculator_rust.so bin/libcalculator_rust.so
echo "  -> bin/libcalculator_rust.so"

echo "== Building calculator server =="
go build -o bin/calculator_server calculator_server/main.go
echo "  -> bin/calculator_server"

echo "== Building generator =="
go build -o bin/generator generator/main.go
echo "  -> bin/generator"

echo "Build complete."
