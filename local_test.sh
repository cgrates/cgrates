#! /usr/bin/env sh
./test.sh
gen=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -local'
go test github.com/cgrates/cgrates/apier/v1 -local
ap1=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -local'
go test github.com/cgrates/cgrates/apier/v2 -local
ap2=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutmysql'
go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutmysql
tpmysql=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutpostgres'
go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutpostgres
tppg=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutmongo'
go test github.com/cgrates/cgrates/apier/v2 -tp -config_dir=tutmongo
tpmongo=$?
echo 'go test github.com/cgrates/cgrates/engine -local -integration'
go test github.com/cgrates/cgrates/engine -local -integration
en=$?
echo 'go test github.com/cgrates/cgrates/cdrc -local'
go test github.com/cgrates/cgrates/cdrc -local
cdrc=$?
echo 'go test github.com/cgrates/cgrates/config -local'
go test github.com/cgrates/cgrates/config -local
cfg=$?
echo 'go test github.com/cgrates/cgrates/utils -local'
go test github.com/cgrates/cgrates/utils -local
utl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -local -integration'
go test github.com/cgrates/cgrates/general_tests -local -integration
gnr=$?
echo 'go test github.com/cgrates/cgrates/agents -integration'
go test github.com/cgrates/cgrates/agents -integration
agts=$?
echo 'go test github.com/cgrates/cgrates/sessionmanager -integration'
go test github.com/cgrates/cgrates/sessionmanager -integration
smg=$?

exit $gen && $ap1 && $ap2 && $tpmysql && $tppg && $tpmongo && $en && $cdrc && $cfg && $utl && $gnr && $agts && $smg
