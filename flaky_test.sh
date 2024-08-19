#!/bin/bash

# This script runs flaky tests on various packages, optionally filtered by dbtype or rpc.
# Usage:
#     - To run all flaky tests, don't add any arguments.
#     - To run flaky tests for gob only, add `-rpc=*gob` as an argument to this script.
#     - To run flaky tests for a single dbtype, add `-dbtype=*mysql` as an argument.
# Example:
# ./flaky_tests.sh -dbtype=*mysql -rpc=*gob

packages=("agents" "apis" "cmd/cgr-console" "cmd/cgr-loader" "dispatchers" "efs" "engine" "ers" "general_tests" "loaders" "registrarc" "sessions")
dbtypes=("*internal" "*mysql" "*mongo" "*postgres")

# Tests that are independent of the dbtype flag and run only once
single_run_packages=("analyzers" "cdrs" "config" "cores" "ees" "utils" "migrator" "services")

results=()

execute_test() {
    local pkg=$1
    shift # Remove the first argument, which is the package name, to pass the rest to go test
    echo "Executing: go test github.com/cgrates/cgrates/$pkg -tags=flaky $@"
    go test "github.com/cgrates/cgrates/$pkg" -tags="flaky" "$@"
    results+=($?)
}

go clean --cache

# Execute flaky tests based on passed arguments
if [ "$#" -ne 0 ]; then
    for pkg in "${packages[@]}"; do
        execute_test "$pkg" "$@"
    done
else
    # Execute flaky tests for all db types if no arguments have been passed
    for db in "${dbtypes[@]}"; do
        for pkg in "${packages[@]}"; do
            execute_test "$pkg" "-dbtype=$db"
        done
    done
fi

# Execute the flaky tests that run only once
for test in "${single_run_packages[@]}"; do
    execute_test "$test"
done

# Check the results and exit with an appropriate code
pass=0
for val in "${results[@]}"; do
    if [ "$val" -ne 0 ]; then
        pass=1
        break
    fi
done
exit $pass
