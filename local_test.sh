#! /usr/bin/env sh

./test.sh
gen=$?
go test github.com/cgrates/cgrates/apier/v1 -local
ap1=$?
go test github.com/cgrates/cgrates/apier/v2 -local
ap2=$?
go test github.com/cgrates/cgrates/engine -local
en=$?
go test github.com/cgrates/cgrates/cdrc -local
cdrc=$?
go test github.com/cgrates/cgrates/config -local
cfg=$?
go test github.com/cgrates/cgrates/utils -local
utl=$?






exit $gen && $ap1 && $ap2 && $en && $cdrc && $cfg && $utl

