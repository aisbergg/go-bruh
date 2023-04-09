#!/usr/bin/env bash
set -eo pipefail

# Install all dev tools.
#
# Usage: bash ./tools/install.sh

pushd "$(dirname "$0")" >/dev/null;

tools=($(grep -ohP '(?<=^\t_ ")[^"]+' tools.go))
for tool in "${tools[@]}"; do
    echo "Installing $tool"
    go install "$tool"
done

popd >/dev/null
