#!/usr/bin/env bash
set -euo pipefail

function headline() {
    echo "================================================================================="
    echo "$@"
    echo "================================================================================="
}


help_text="$0 [OPTIONS] COMMAND ARGUMENTS

Run benchmarks and compare the results with the run with the given name or the last run.

Options:
  -h|--help      Show this help message and quite
  -d|--dir=./benchmarks    Directory containing the benchmarks
  -c|--count=6   Number of times to run the benchmarks
  -u|--cpu=      Number of CPUs to use

Commands:
  bench                    Run the benchmarks
  bench_compare_files      Run the benchmarks and compare files afterwards
  bench_compare_functions  Run the benchmarks and compare functions afterwards
  compare_files            Compare results of two or more files
  compare_functions        Compare functions inside a file

Arguments:
  bench BENCHMARK_EXPR RUN_NAME
  bench_compare_files BENCHMARK_EXPR RUN_NAME COMPARE_RUN_NAME...
  bench_compare_functions BENCHMARK_EXPR RUN_NAME
  compare_files BENCHMARK_EXPR RUN_NAME COMPARE_RUN_NAME...
  compare_functions BENCHMARK_EXPR RUN_NAME"

args=$(getopt -o "h,c:,u:" -l "help,count:,cpu:" -- "$@")
eval set -- "$args"

BENCHMARK_DIR='./benchmarks'
CPUS=
COUNT=6
while [[ $# > 0 ]]; do
    case "$1" in
        -h|--help) echo "$help_text"; exit 0;;
        -d|--dir) BENCHMARK_DIR="$2"; shift 2;;
        -c|--count) COUNT="$2"; shift 2;;
        -u|--cpu) CPUS="$2"; shift 2;;
        --) shift ; break ;;
        *) echo "Unknown option '$1'"; exit 2;;
    esac
done
if [[ $# < 3 ]]; then
    echo "Missing arguments"
    echo "$help_text"
    exit 2
fi
COMMAND="$1"
BENCHMARK_EXPR="$2"
RUN_NAME="$3"
shift 3

# setup test directory
benchmark_results_path="$(dirname "$0")/results"
mkdir -p "$benchmark_results_path"

benchmark_name=$(echo "$BENCHMARK_EXPR" | tr -cd '[:alnum:]')
if [[ -n "$benchmark_name" ]]; then
    benchmark_name="${benchmark_name}__"
fi
new_results_file="$benchmark_results_path/${benchmark_name}${RUN_NAME}.txt"


function bench() {
    headline "Running benchmarks"
    go test -run='^$' -bench="$BENCHMARK_EXPR" -benchmem -count="$COUNT" -cpu="$CPUS" "$BENCHMARK_DIR" | tee "$new_results_file"
}

function compare_files() {
    # $1: benchmark_name
    # $2: run_name
    # $@: compare_run_names
    local benchmark_name="$1"
    local run_name="$2"
    shift 2
    local compare_run_names=("$@")

    for compare_run_name in "${compare_run_names[@]}"; do
        compare_results_file="$benchmark_results_path/${benchmark_name}${compare_run_name}.txt"
        if [[ ! -f "$compare_results_file" ]]; then
            echo "Compare results file '$compare_results_file' not found."
            exit 2
        fi
    done

    headline "Comparing files"
    args=()
    for compare_run_name in "${compare_run_names[@]}"; do
        args+=("${compare_run_name}=${benchmark_results_path}/${benchmark_name}${compare_run_name}.txt")
    done
    benchstat "$run_name=$new_results_file" "${args[@]}"
}

function compare_functions() {
    # $1: benchmark_name
    # $2: run_name
    local benchmark_name="$1"
    local run_name="$2"

    headline "Comparing functions"
    benchstat -col .name -row "" "$new_results_file"
}


if [[ "$COMMAND" == "compare_files" ]]; then
    if [[ $# < 1 ]]; then
        echo "Missing compare run names"
        echo "$help_text"
        exit 2
    fi
    compare_files "$benchmark_name" "$RUN_NAME" "$@"

elif [[ "$COMMAND" == "compare_functions" ]]; then
    compare_functions "$benchmark_name" "$RUN_NAME"

elif [[ "$COMMAND" == "bench" ]]; then
    bench

elif [[ "$COMMAND" == "bench_compare_files" ]]; then
    if [[ $# < 1 ]]; then
        echo "Missing compare run names"
        echo "$help_text"
        exit 2
    fi
    bench
    compare_files "$benchmark_name" "$RUN_NAME" "$@"

elif [[ "$COMMAND" == "bench_compare_functions" ]]; then
    bench
    compare_functions "$benchmark_name" "$RUN_NAME"

else
    echo "Unknown command '$COMMAND'"
    echo "$help_text"
    exit 2
fi
