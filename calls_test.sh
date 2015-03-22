#! /usr/bin/env sh

./local_test.sh
lcl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -calls'
go test github.com/cgrates/cgrates/general_tests -calls
gnr=$?

exit $gen && $gnr