#! /usr/bin/env sh

./test.sh
gen=$?
go test github.com/cgrates/cgrates/apier -local
ap=$?
go test github.com/cgrates/cgrates/engine -local
en=$?
go test github.com/cgrates/cgrates/cdrc -local
cdrc=$?
go test github.com/cgrates/cgrates/mediator -local
med=$?




exit $gen && $ap && $en && $cdrc && $med

