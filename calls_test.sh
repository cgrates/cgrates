#! /usr/bin/env sh

# ./local_test.sh
# lcl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestFreeswitchCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestFreeswitchCalls
gnr1=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestAsteriskCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestAsteriskCalls
gnr2=$?

exit $gnr1 && $gnr2