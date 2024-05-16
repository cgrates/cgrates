#!/bin/bash

# This script is used to run integration tests on various packages with different tags and dbtypes.
# Usage:
#     - To run all the integration tests, don't add any arguments.
#     - To run the integration tests for gob only add `-rpc=*gob` as an argument to this script.
#     - To run for a single dbtype add `-dbtype=*mysql` as an argument.
# Example:
# ./integration_test.sh -dbtype=*mysql -rpc=*gob

packages=("agents" "apier/v1" "apier/v2" "dispatchers" "engine" "ers" "loaders" "general_tests" "sessions")
dbtypes=("*internal" "*mysql" "*mongo" "*postgres")

# Tests that are independent of the dbtype flag and run only once
single_run_packages=("config" "services" "migrator")

results=()

execute_test() {
    echo "Executing: go test github.com/cgrates/cgrates/$1 -tags=$2 $3"
    go test "github.com/cgrates/cgrates/$1" -tags="$2" "$3"
    results+=($?)
}

go clean --cache 

# Execute tests based on passed arguments
if [ "$#" -ne 0 ]; then
    for pkg in "${packages[@]}"; do
        execute_test "$pkg" "integration" "$@"
        [ "$pkg" == "apier/v1" ] && execute_test "$pkg" "offline" "$@"
    done
else
   # Execute tests for all db types if no arguments have been passed
    for db in "${dbtypes[@]}"; do
        for pkg in "${packages[@]}"; do
            execute_test "$pkg" "integration" "-dbtype=$db"
            
            # Temporarily skip offline tests for dbtype=*internal until fixed
            if [ "$db" != "*internal" ] && [ "$pkg" == "apier/v1" ]; then
                execute_test "$pkg" "offline" "-dbtype=$db"
            fi
        done
    done
fi

# Execute the tests that run only once
for test in "${single_run_packages[@]}"; do
    execute_test "$test" "integration"
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