#!/usr/bin/env bash
set -euo pipefail

# detect-secrets is a tool to detect secrets in the repository. Unfortunately it
# does not have a "check" mode, so we have to do some manual work to compare the
# known secrets with the newly detected secrets.

tmp_baseline=$(mktemp)
trap 'rm -f "$tmp_baseline"' EXIT

cp .secrets.baseline "$tmp_baseline"
detect-secrets scan --baseline "$tmp_baseline" $(find . -type f ! -name '.secrets.*' ! -path '*/.git*')

# if there is any difference between the known and newly detected secrets, break the build
list_secret_ids() {
    jq -r '.results | to_entries[] | .key as $file | .value[] | [$file, (.line_number | tostring), .type, .hashed_secret] | @tsv' "$1" | sort
}

list_secret_locations() {
    jq -r '.results | to_entries[] | .key as $file | .value[] | "\($file):\(.line_number) [\(.type)]"' "$1" | sort
}

if diff <(list_secret_ids .secrets.baseline) <(list_secret_ids "$tmp_baseline") >/dev/null; then
    echo "No new secrets detected"
    exit 0
fi

echo "Detected new secrets in the repo" >&2
comm -13 <(list_secret_ids .secrets.baseline) <(list_secret_ids "$tmp_baseline") \
    | cut -f1-3 \
    | awk -F '\t' '{print "  at " $1 ":" $2}' >&2
exit 1
