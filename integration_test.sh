#! /usr/bin/env sh
go clean --cache
./test.sh
gen=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration
ap1=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration
ap2=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration'
go test github.com/cgrates/cgrates/engine -tags=integration
en=$?
echo 'go test github.com/cgrates/cgrates/cdrc -tags=integration'
go test github.com/cgrates/cgrates/cdrc -tags=integration
cdrc=$?
echo 'go test github.com/cgrates/cgrates/config -tags=integration'
go test github.com/cgrates/cgrates/config -tags=integration
cfg=$?
echo 'go test github.com/cgrates/cgrates/utils -tags=integration'
go test github.com/cgrates/cgrates/utils -tags=integration
utl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration'
go test github.com/cgrates/cgrates/general_tests -tags=integration
gnr=$?
echo 'go test github.com/cgrates/cgrates/agents -tags=integration'
go test github.com/cgrates/cgrates/agents -tags=integration
agts=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration'
go test github.com/cgrates/cgrates/sessions -tags=integration
smg=$?
echo 'go test github.com/cgrates/cgrates/migrator -tags=integration'
go test github.com/cgrates/cgrates/migrator -tags=integration
mgr=$?
echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration'
go test github.com/cgrates/cgrates/dispatchers -tags=integration
dis=$?

exit $gen && $ap1 && $ap2 && $en && $cdrc && $cfg && $utl && $gnr && $agts && $smg && $mgr && $dis
