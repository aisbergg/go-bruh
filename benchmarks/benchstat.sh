#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

function headline() {
    echo "======================================================================"
    echo "$@"
    echo "======================================================================"
}

new_results="./results/$(date +"%Y-%m-%dT%H:%M:%S").txt"
if [[ "$#" > 0 && -n "$1" ]]; then
    previous_results="./results/$1"
else
    previous_results="$(find ./results -maxdepth 1 -name "*.txt" | tail -n1)"
fi

if [[ -n "$previous_results" && ! -f "$previous_results" ]]; then
    echo "Previous results file not found: $previous_results"
    exit 1
fi

headline "Running benchmarks"
mkdir -p results
go test -bench=. -benchmem -count=6 | tee "$new_results"

if [[ -z "$previous_results" ]]; then
    echo
    echo "No previous results found, cannot compare; exiting"
    exit 0
fi

headline "Comparing results"
benchstat "$previous_results" "$new_results"
