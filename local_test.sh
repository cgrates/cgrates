#! /usr/bin/env sh
./test.sh
gen=$?
go test -local -integration $(glide novendor)
exit $gen && $?
