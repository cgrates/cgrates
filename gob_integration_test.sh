#! /usr/bin/env sh
go clean --cache
./test.sh
gen=$?

# Internal
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*internal -rpc=*gob
ap1_internal=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*internal -rpc=*gob
ap2_internal=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*internal -rpc=*gob
en_internal=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*internal -rpc=*gob
ers_internal=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*internal -rpc=*gob
lds_internal=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*internal -rpc=*gob
gnr_internal=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*internal -rpc=*gob
agts_internal=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*internal -rpc=*gob
smg_internal=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*internal -rpc=*gob
dis_internal=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*internal -rpc=*gob
offline_internal=$?
# SQL
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mysql -rpc=*gob
ap1_mysql=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mysql -rpc=*gob
ap2_mysql=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mysql -rpc=*gob
en_mysql=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mysql -rpc=*gob
ers_mysql=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mysql -rpc=*gob
lds_mysql=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql -rpc=*gob
gnr_mysql=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mysql -rpc=*gob
agts_mysql=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mysql -rpc=*gob
smg_mysql=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mysql -rpc=*gob
dis_mysql=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mysql -rpc=*gob
offline_mysql=$?
# Mongo
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*mongo -rpc=*gob
ap1_mongo=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*mongo -rpc=*gob
ap2_mongo=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*mongo -rpc=*gob
en_mongo=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*mongo -rpc=*gob
ers_mongo=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*mongo -rpc=*gob
lds_mongo=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mongo -rpc=*gob
gnr_mongo=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*mongo -rpc=*gob
agts_mongo=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*mongo -rpc=*gob
smg_mongo=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*mongo -rpc=*gob
dis_mongo=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*mongo -rpc=*gob
offline_mongo=$?
# Postgres
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -dbtype=*postgres -rpc=*gob
ap1_postgres=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -dbtype=*postgres -rpc=*gob
ap2_postgres=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/engine -tags=integration -dbtype=*postgres -rpc=*gob
en_postgres=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/ers -tags=integration -dbtype=*postgres -rpc=*gob
ers_postgres=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/loaders -tags=integration -dbtype=*postgres -rpc=*gob
lds_postgres=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*postgres -rpc=*gob
gnr_postgres=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/agents -tags=integration -dbtype=*postgres -rpc=*gob
agts_postgres=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/sessions -tags=integration -dbtype=*postgres -rpc=*gob
smg_postgres=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/dispatchers -tags=integration -dbtype=*postgres -rpc=*gob
dis_postgres=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline -dbtype=*postgres -rpc=*gob
offline_postgres=$?

echo 'go test github.com/cgrates/cgrates/config -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/config -tags=integration -rpc=*gob
cfg=$?
echo 'go test github.com/cgrates/cgrates/migrator -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/migrator -tags=integration -rpc=*gob
mgr=$?
echo 'go test github.com/cgrates/cgrates/services -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/services -tags=integration -rpc=*gob
srv=$?

exit $gen && $ap1_internal && $ap2_internal && $en_internal && $ers_internal && $lds_internal && 
$gnr_internal && $agts_internal && $smg_internal && $dis_internal && $offline_internal && $ap1_mysql && 
$ap2_mysql && $en_mysql && $ers_mysql && $lds_mysql && $gnr_mysql && $agts_mysql && $smg_mysql && 
$dis_mysql && $offline_mysql && $ap1_mongo && $ap2_mongo && $en_mongo && $ers_mongo && $lds_mongo && 
$gnr_mongo && $agts_mongo && $smg_mongo && $dis_mongo && $offline_mongo && $ap1_postgres && 
$ap2_postgres && $en_postgres && $ers_postgres && $lds_postgres && $gnr_postgres && $agts_postgres && 
$smg_postgres && $dis_postgres && $offline_postgres && $cfg && $mgr && $srv