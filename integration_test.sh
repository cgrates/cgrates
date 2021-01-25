#!/bin/bash
set -e
go clean --cache
results=()
./test.sh
results+=($?)
if [ "$#" -ne 0 ]; then
# to run the integration tests for gob only add `-rpc=*gob` as argument to this script
# to run for a single dbtype add `-dbtype=*mysql` as argument
# ./integaration_tes.sh -dbtype=*mysql -rpc=*gob
echo "go test github.com/cgrates/cgrates/apier/v1 -tags=integration $@"
go test github.com/cgrates/cgrates/apier/v1 -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/apier/v2 -tags=integration $@"
go test github.com/cgrates/cgrates/apier/v2 -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/engine  -tags=integration $@"
go test github.com/cgrates/cgrates/engine -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/ers -tags=integration $@"
go test github.com/cgrates/cgrates/ers -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/loaders -tags=integration $@"
go test github.com/cgrates/cgrates/loaders -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/general_tests -tags=integration $@"
go test github.com/cgrates/cgrates/general_tests -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/agents -tags=integration $@"
go test github.com/cgrates/cgrates/agents -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/sessions -tags=integration $@"
go test github.com/cgrates/cgrates/sessions -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatchers -tags=integration $@"
go test github.com/cgrates/cgrates/dispatchers -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatcherh -tags=integration $@"
go test github.com/cgrates/cgrates/dispatcherh -tags=integration $@
results+=($?)
echo "go test github.com/cgrates/cgrates/apier/v1 -tags=offline $@"
go test github.com/cgrates/cgrates/apier/v1 -tags=offline $@
results+=($?)
echo "go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration $@"
go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration $@
results+=($?)
else
# Internal
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*internal"
go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*internal
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal
results+=($?)
echo "go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*internal"
go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*internal
results+=($?)
# SQL
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*mysql"
go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*mysql
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql
results+=($?)
echo "go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*mysql"
go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*mysql
results+=($?)
# Mongo
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*mongo"
go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*mongo
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo
results+=($?)
echo "go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*mongo"
go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*mongo
results+=($?)
# Postgres
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres
results+=($?)
echo "go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*postgres"
go test github.com/cgrates/cgrates/dispatcherh -tags=integration -dbtype=*postgres
results+=($?)
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres
results+=($?)
echo "go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*postgres"
go test github.com/cgrates/cgrates/cmd/cgr-loader -tags=integration -dbtype=*postgres
results+=($?)

fi

echo "go test github.com/cgrates/cgrates/analyzers -tags=integration"
go test github.com/cgrates/cgrates/analyzers -tags=integration
results+=($?)
echo "go test github.com/cgrates/cgrates/ees -tags=integration"
go test github.com/cgrates/cgrates/ees -tags=integration
results+=($?)
echo 'go test github.com/cgrates/cgrates/config -tags=integration'
go test github.com/cgrates/cgrates/config -tags=integration
results+=($?)
echo 'go test github.com/cgrates/cgrates/utils -tags=integration'
go test github.com/cgrates/cgrates/utils -tags=integration
results+=($?)
echo 'go test github.com/cgrates/cgrates/migrator -tags=integration'
go test github.com/cgrates/cgrates/migrator -tags=integration
results+=($?)
echo 'go test github.com/cgrates/cgrates/services -tags=integration'
go test github.com/cgrates/cgrates/services -tags=integration
results+=($?)


pass=1
for val in ${results[@]}; do
   (( pass=$pass&&$val))
done
exit $pass