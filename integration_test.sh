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
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal
offline_internal=$?
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
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql
offline_mysql=$?
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
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo
offline_mongo=$?
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
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres
offline_postgres=$?

echo 'go test github.com/cgrates/cgrates/config -tags=integration'
go test github.com/cgrates/cgrates/config -tags=integration
cfg=$?
echo 'go test github.com/cgrates/cgrates/migrator -tags=integration'
go test github.com/cgrates/cgrates/migrator -tags=integration
mgr=$?
echo 'go test github.com/cgrates/cgrates/services -tags=integration'
go test github.com/cgrates/cgrates/services -tags=integration
srv=$?

exit $gen && $ap1_internal && $ap2_internal && $en_internal && $ers_internal && $lds_internal && 
$gnr_internal && $agts_internal && $smg_internal && $dis_internal && $offline_internal && $ap1_mysql && 
$ap2_mysql && $en_mysql && $ers_mysql && $lds_mysql && $gnr_mysql && $agts_mysql && $smg_mysql && 
$dis_mysql && $offline_mysql && $ap1_mongo && $ap2_mongo && $en_mongo && $ers_mongo && $lds_mongo && 
$gnr_mongo && $agts_mongo && $smg_mongo && $dis_mongo && $offline_mongo && $ap1_postgres && 
$ap2_postgres && $en_postgres && $ers_postgres && $lds_postgres && $gnr_postgres && $agts_postgres && 
$smg_postgres && $dis_postgres && $offline_postgres && $cfg && $mgr && $srv