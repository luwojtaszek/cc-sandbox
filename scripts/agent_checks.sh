#!/usr/bin/env bash
# agent_checks.sh - Context-efficient checks for AI agents
# Inspired by: https://www.hlyr.dev/blog/context-efficient-backpressure
#
# Runs: build → test → lint
# On success: shows ✓ with summary
# On failure: shows only relevant output

set -e

# Change to repo root (script is in scripts/ directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Tool versions (keep in sync with Makefile and lefthook.yml)
GOLANGCI_LINT_VERSION="v1.64.8"

run_check() {
    local name="$1"
    local cmd="$2"
    local tmp_file
    tmp_file=$(mktemp)

    pushd cli > /dev/null
    if eval "$cmd" > "$tmp_file" 2>&1; then
        popd > /dev/null
        echo "✓ $name"
        rm -f "$tmp_file"
        return 0
    else
        popd > /dev/null
        echo "✗ $name"
        echo "Command: cd cli && $cmd"
        echo "---"
        cat "$tmp_file"
        rm -f "$tmp_file"
        return 1
    fi
}

# Go test with filtered output - shows only failures
run_go_test() {
    local tmp_file
    tmp_file=$(mktemp)

    pushd cli > /dev/null
    if go test ./... > "$tmp_file" 2>&1; then
        popd > /dev/null
        # Extract pass count from go test output
        local count
        count=$(grep -oE "^ok.*[0-9]+\.[0-9]+s" "$tmp_file" | wc -l | tr -d ' ')
        if [ -n "$count" ] && [ "$count" -gt 0 ]; then
            echo "✓ test ($count packages)"
        else
            echo "✓ test"
        fi
        rm -f "$tmp_file"
        return 0
    else
        popd > /dev/null
        echo "✗ test"
        echo "Command: cd cli && go test ./..."
        echo "---"
        # Filter to show only failures - Go test output is already fairly clean
        # Show FAIL lines, error messages, and summary
        grep -E "(^---|FAIL|Error|panic|expected|got)" "$tmp_file" || cat "$tmp_file"
        rm -f "$tmp_file"
        return 1
    fi
}

echo "Running agent checks..."
run_check "build" "go build ./..."
run_go_test
run_check "lint" "go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run --fix"
echo "All checks passed ✓"
