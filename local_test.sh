#! /usr/bin/env sh

./test.sh
gen=$?
go test github.com/cgrates/cgrates/apier/v1 -local
ap=$?
go test github.com/cgrates/cgrates/engine -local
en=$?
go test github.com/cgrates/cgrates/cdrc -local
cdrc=$?




exit $gen && $ap && $en && cdrc

