#! /usr/bin/env sh

# ./local_test.sh
lcl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call'
go test github.com/cgrates/cgrates/general_tests -tags=call -timeout=10h
gnr=$?

exit $gen && $gnr