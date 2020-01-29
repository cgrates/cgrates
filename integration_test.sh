#! /usr/bin/env sh
go clean --cache
./test.sh
gen=$?

# Internal
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal
ap1_internal=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal
ap2_internal=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*internal
en_internal=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal
ers_internal=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal
lds_internal=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal
gnr_internal=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal
agts_internal=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal
smg_internal=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal
dis_internal=$?
# SQL
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql
ap1_mysql=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql
ap2_mysql=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mysql
en_mysql=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql
ers_mysql=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql
lds_mysql=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql
gnr_mysql=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql
agts_mysql=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql
smg_mysql=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql
dis_mysql=$?
# Mongo
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo
ap1_mongo=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo
ap2_mongo=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mongo
en_mongo=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo
ers_mongo=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo
lds_mongo=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo
gnr_mongo=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo
agts_mongo=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo
smg_mongo=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo
dis_mongo=$?
# Postgres
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres
ap1_postgres=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres
ap2_postgres=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*postgres
en_postgres=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres
ers_postgres=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres
lds_postgres=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres
gnr_postgres=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres
agts_postgres=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres
smg_postgres=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres
dis_postgres=$?

echo 'go test github.com/cgrates/cgrates/config -tags=integration'
go test github.com/cgrates/cgrates/config -tags=integration
cfg=$?
echo 'go test github.com/cgrates/cgrates/migrator -tags=integration'
go test github.com/cgrates/cgrates/migrator -tags=integration
mgr=$?
echo 'go test github.com/cgrates/cgrates/services -tags=integration'
go test github.com/cgrates/cgrates/services -tags=integration
srv=$?
#All

echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline
offline=$?
# to do: add '&& $ap1_internal' 
exit $gen && $ap1_sql && $ap1_mongo && $ap2 && $en && $cfg && $utl && $gnr && $agts && $smg && $mgr && $dis && $lds && $ers && $srv && $offline
