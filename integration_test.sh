#!/bin/bash
go clean --cache

# Define directories to test and dbtypes
directories=("apier/v1" "apier/v2" "engine" "ers" "loaders" "general_tests" "agents" "sessions" "dispatchers")
dbtypes=("internal" "mysql" "mongo" "postgres")

results=()

# check if any arguments passed
if [ "$#" -ne 0 ]; then
    for directory in "${directories[@]}"; do
        command="go test github.com/cgrates/cgrates/$directory -tags=integration $@"
        echo $command
        $command
        results+=($?)
    done
else
    # No arguments passed, running with predefined options
    for dbtype in "${dbtypes[@]}"; do
        for directory in "${directories[@]}"; do
            command="go test github.com/cgrates/cgrates/$directory -tags=integration -dbtype=*${dbtype}"
            echo $command
            $command
            results+=($?)
        done
    done
fi

# Run tests for packages that don't rely on db connections
directories=("config" "migrator" "services")
for directory in "${directories[@]}"; do
    command="go test github.com/cgrates/cgrates/$directory -tags=integration"
    echo $command
    $command
    results+=($?)
done

# Check if any tests failed
pass=1
for val in ${results[@]}; do
    (( pass=$pass||$val))
done
exit $pass
